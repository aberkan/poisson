package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/zeace/poisson/analyzer"
	"github.com/zeace/poisson/fetcher"
)

func main() {
	var (
		apiKey  = flag.String("api-key", "", "OpenAI API key (or set OPENAI_API_KEY environment variable)")
		verbose = flag.Bool("verbose", false, "Show verbose output")
		url     = flag.String("url", "", "URL of the article to analyze")
	)
	flag.Parse()

	if *url == "" {
		fmt.Fprintf(os.Stderr, "Error: URL required\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Get API key from flag, secrets file, or environment
	apiKeyValue := *apiKey
	if apiKeyValue == "" {
		// Try reading from secrets file
		secretKey, err := os.ReadFile("secrets/openai_key")
		if err == nil {
			apiKeyValue = strings.TrimSpace(string(secretKey))
		} else {
			// Fall back to environment variable
			apiKeyValue = os.Getenv("OPENAI_API_KEY")
		}
	}

	fmt.Printf("Fetching article from: %s\n", *url)
	content, err := fetcher.FetchArticleContent(*url, *verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		preview := content
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		fmt.Printf("\nFetched %d characters of content\n", len(content))
		fmt.Printf("Preview: %s\n\n", preview)
	}

	fmt.Println("Analyzing content with LLM...")
	analysis, err := analyzer.AnalyzeWithLLM(content, apiKeyValue)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("ANALYSIS RESULTS")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println(analysis)
	fmt.Println(strings.Repeat("=", 60))
}

