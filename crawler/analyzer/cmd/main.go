package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/zeace/poisson/crawler/analyzer"
)

const maxContentLength = 8000

func main() {
	var (
		apiKey   = flag.String("api-key", "", "OpenAI API key (or set OPENAI_API_KEY environment variable)")
		filePath = flag.String("file", "", "Path to the file containing article content")
	)
	flag.Parse()

	if *filePath == "" {
		fmt.Fprintf(os.Stderr, "Error: file path required\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Get API key from flag or environment
	apiKeyValue := *apiKey

	// Read content from file
	fmt.Printf("Reading content from: %s\n", *filePath)
	content, err := os.ReadFile(*filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	contentStr := string(content)
	fmt.Printf("Read %d characters from file\n", len(contentStr))
	fmt.Println("Analyzing content with LLM...")

	// Truncate content if too long
	truncatedContent := contentStr
	if len(truncatedContent) > maxContentLength {
		truncatedContent = truncatedContent[:maxContentLength] + "... [content truncated]"
	}
	prompt := analyzer.AddBodyToPrompt(analyzer.JokePromptTemplate, truncatedContent)
	analysis, err := analyzer.AnalyzeWithLLM(prompt, apiKeyValue)
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
