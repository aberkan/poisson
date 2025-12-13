package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/datastore"
	"github.com/zeace/poisson/crawler/analyzer"
	"github.com/zeace/poisson/crawler/fetcher"
	"github.com/zeace/poisson/crawler/rssfetcher"
	"github.com/zeace/poisson/models"
)

// displayAnalysis displays the analysis results and related information.
// If verbose is true, it shows a preview of the content.
// If articleNum and totalArticles are provided (> 0), it shows article progress.
func displayAnalysis(
	analysis *models.AnalysisResult, title, content string,
	verbose bool,
	articleNum, totalArticles int,
) {
	// Show article progress if provided
	if articleNum > 0 && totalArticles > 0 {
		fmt.Printf("Article %d/%d\n", articleNum, totalArticles)
		fmt.Println(strings.Repeat("-", 60))
	}

	// Show verbose preview if requested
	if verbose {
		fmt.Printf("Title: %s\n", title)
		preview := content
		previewLen := 200
		if len(preview) > previewLen {
			preview = preview[:previewLen] + "..."
		}
		fmt.Printf("Preview: %s\n\n", preview)
	}

	// Show analyzing message
	fmt.Println("Analyzing content with LLM...")

	// Display results
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("ANALYSIS RESULTS")
	fmt.Println(strings.Repeat("=", 60))
	if analysis.JokePercentage == nil {
		fmt.Printf("Joke Percentage: null (no mention of jokes)\n")
	} else {
		fmt.Printf("Joke Percentage: %d\n", *analysis.JokePercentage)
	}
	if analysis.JokeReasoning != nil {
		fmt.Printf("Joke Reasoning: %s\n", *analysis.JokeReasoning)
	}
	fmt.Println(strings.Repeat("=", 60))
}

func main() {
	var (
		apiKey  = flag.String("api-key", "", "OpenAI API key (or set OPENAI_API_KEY environment variable)")
		verbose = flag.Bool("verbose", false, "Show verbose output")
		url     = flag.String("url", "", "URL of the article to analyze")
		rss     = flag.String("rss", "", "URL of the RSS feed to analyze")
		max     = flag.Int("max", 5, "Maximum number of articles to fetch from RSS feed")
		mode    = flag.String("mode", "joke", "Analysis mode (joke)")
	)
	flag.Parse()

	// Validate mode
	promptMode, err := analyzer.VerifyValidMode(*mode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: unknown mode '%s'. Valid modes: joke\n", *mode)
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

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

	// Set up Datastore client
	ctx := context.Background()
	projectID := "poisson-berkan"
	datastoreClient, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer datastoreClient.Close()

	if urlProvided {
		// Single URL mode - existing behavior
		fmt.Printf("Fetching article from: %s\n", *url)
		page, cachePath, err := fetcher.FetchArticleContent(ctx, *url, *verbose, datastoreClient)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		_ = cachePath // cache path available for future use

		analysis, err := analyzer.Analyze(page, apiKeyValue, promptMode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		displayAnalysis(analysis, page.Title, page.Content, *verbose, 0, 0)
	} else {
		// RSS mode - fetch articles and analyze each
		pages, err := rssfetcher.FetchRSSArticles(ctx, *rss, *max, *verbose, datastoreClient)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching RSS articles: %v\n", err)
			os.Exit(1)
		}

		if len(pages) == 0 {
			fmt.Fprintf(os.Stderr, "Error: no articles fetched from RSS feed\n")
			os.Exit(1)
		}

		fmt.Printf("\n%s\n", strings.Repeat("=", 60))
		fmt.Printf("Analyzing %d article(s) from RSS feed\n", len(pages))
		fmt.Printf("%s\n\n", strings.Repeat("=", 60))

		for i, page := range pages {
			showSeparator := i < len(pages)-1
			analysis, err := analyzer.Analyze(page, apiKeyValue, promptMode)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error analyzing article %d: %v\n", i+1, err)
				fmt.Println(strings.Repeat("-", 60))
				if showSeparator {
					fmt.Println()
				}
				continue
			}
			displayAnalysis(analysis, page.Title, page.Content, *verbose, i+1, len(pages))
		}
	}
}
