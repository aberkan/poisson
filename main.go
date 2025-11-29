package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/zeace/poisson/analyzer"
	"github.com/zeace/poisson/fetcher"
	"github.com/zeace/poisson/rssfetcher"
)

// analyzeAndDisplay analyzes content with LLM and displays the results.
// If verbose is true, it shows a preview of the content.
// If articleNum and totalArticles are provided (> 0), it shows article progress.
// If showSeparator is true, it shows a separator after the results.
func analyzeAndDisplay(content, apiKey string, verbose bool, articleNum, totalArticles int, showSeparator bool) error {
	// Show article progress if provided
	if articleNum > 0 && totalArticles > 0 {
		fmt.Printf("Article %d/%d\n", articleNum, totalArticles)
		fmt.Println(strings.Repeat("-", 60))
	}

	// Show verbose preview if requested
	if verbose {
		preview := content
		previewLen := 200
		if len(preview) > previewLen {
			preview = preview[:previewLen] + "..."
		}
		if articleNum > 0 {
			fmt.Printf("Content length: %d characters\n", len(content))
			fmt.Printf("Preview: %s\n\n", preview)
		} else {
			fmt.Printf("\nFetched %d characters of content\n", len(content))
			fmt.Printf("Preview: %s\n\n", preview)
		}
	}

	// Analyze content
	fmt.Println("Analyzing content with LLM...")
	analysis, err := analyzer.AnalyzeWithLLM(content, apiKey)
	if err != nil {
		return fmt.Errorf("error analyzing content: %w", err)
	}

	// Display results
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("ANALYSIS RESULTS")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println(analysis)
	fmt.Println(strings.Repeat("=", 60))
	if showSeparator {
		fmt.Println(strings.Repeat("-", 60))
	}

	return nil
}

func main() {
	var (
		apiKey  = flag.String("api-key", "", "OpenAI API key (or set OPENAI_API_KEY environment variable)")
		verbose = flag.Bool("verbose", false, "Show verbose output")
		url     = flag.String("url", "", "URL of the article to analyze")
		rss     = flag.String("rss", "", "URL of the RSS feed to analyze")
		max     = flag.Int("max", 5, "Maximum number of articles to fetch from RSS feed")
	)
	flag.Parse()

	// Validate that exactly one of --url or --rss is provided
	urlProvided := *url != ""
	rssProvided := *rss != ""

	if !urlProvided && !rssProvided {
		fmt.Fprintf(os.Stderr, "Error: exactly one of --url or --rss must be provided\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	if urlProvided && rssProvided {
		fmt.Fprintf(os.Stderr, "Error: cannot specify both --url and --rss\n")
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

	if urlProvided {
		// Single URL mode - existing behavior
		fmt.Printf("Fetching article from: %s\n", *url)
		content, err := fetcher.FetchArticleContent(*url, *verbose)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if err := analyzeAndDisplay(content, apiKeyValue, *verbose, 0, 0, false); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		// RSS mode - fetch articles and analyze each
		articles, err := rssfetcher.FetchRSSArticles(*rss, *max, *verbose)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching RSS articles: %v\n", err)
			os.Exit(1)
		}

		if len(articles) == 0 {
			fmt.Fprintf(os.Stderr, "Error: no articles fetched from RSS feed\n")
			os.Exit(1)
		}

		fmt.Printf("\n%s\n", strings.Repeat("=", 60))
		fmt.Printf("Analyzing %d article(s) from RSS feed\n", len(articles))
		fmt.Printf("%s\n\n", strings.Repeat("=", 60))

		for i, article := range articles {
			showSeparator := i < len(articles)-1
			if err := analyzeAndDisplay(article, apiKeyValue, *verbose, i+1, len(articles), showSeparator); err != nil {
				fmt.Fprintf(os.Stderr, "Error analyzing article %d: %v\n", i+1, err)
				fmt.Println(strings.Repeat("-", 60))
				if showSeparator {
					fmt.Println()
				}
				continue
			}
		}
	}
}

