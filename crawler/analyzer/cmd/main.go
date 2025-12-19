package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/zeace/poisson/crawler/analyzer"
)

func main() {
	var (
		apiKey   = flag.String("api-key", "", "OpenAI API key (or set OPENAI_API_KEY environment variable)")
		filePath = flag.String("file", "", "Path to the file containing article content")
		mode     = flag.String("mode", "joke", "Analysis mode (joke)")
	)
	flag.Parse()

	// Validate mode
	promptMode, err := analyzer.VerifyValidMode(*mode)
	if err != nil {
		log.Printf("Error: unknown mode '%s'. Valid modes: joke\n", *mode)
		log.Printf("Usage: %s [flags]\n", os.Args[0])
		flag.PrintDefaults()
		log.Fatalf("")
	}

	if *filePath == "" {
		log.Printf("Error: file path required\n")
		log.Printf("Usage: %s [flags]\n", os.Args[0])
		flag.PrintDefaults()
		log.Fatalf("")
	}

	// Get API key from flag or environment
	apiKeyValue := *apiKey

	// Read content from file
	log.Printf("Reading content from: %s\n", *filePath)
	content, err := os.ReadFile(*filePath)
	if err != nil {
		log.Fatalf("Error reading file: %v\n", err)
	}

	contentStr := string(content)
	log.Printf("Read %d characters from file\n", len(contentStr))
	log.Printf("Analyzing content with LLM...\n")

	// For file-based analyzer, no title is available - use empty string
	prompt, err := analyzer.GeneratePrompt(promptMode, "", contentStr)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
	analysis, err := analyzer.AnalyzeWithLLM(prompt, apiKeyValue)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	log.Printf("\n%s\n", strings.Repeat("=", 60))
	log.Printf("ANALYSIS RESULTS\n")
	log.Printf("%s\n", strings.Repeat("=", 60))
	log.Printf("%s\n", analysis)
	log.Printf("%s\n", strings.Repeat("=", 60))
}
