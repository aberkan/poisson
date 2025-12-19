package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/zeace/poisson/crawler/analyzer"
	"github.com/zeace/poisson/crawler/fetcher"
	"github.com/zeace/poisson/crawler/rssfetcher"
	"github.com/zeace/poisson/lib"
	"github.com/zeace/poisson/models"
)

// displayAnalysis displays the analysis results and related information.
// If verbose is true, it shows a preview of the content.
// If articleNum and totalArticles are provided (> 0), it shows article progress.
func displayAnalysis(
	analysis *models.AnalysisResult, title, url, content string,
	verbose bool,
	articleNum, totalArticles int,
) {
	// Show article progress if provided
	if articleNum > 0 && totalArticles > 0 {
		log.Printf("\n")
		log.Printf("%s\n", strings.Repeat("-", 120))
		log.Printf("Article %d/%d\n", articleNum, totalArticles)
	}

	// Show verbose preview if requested
	if verbose {
		log.Printf("Title: %s\n", title)
		log.Printf("URL: %s\n", url)
		preview := content
		previewLen := 200
		if len(preview) > previewLen {
			preview = preview[:previewLen] + "..."
		}
		log.Printf("Preview: %s\n\n", preview)
	}

	// Show analyzing message
	log.Printf("Analyzing content with LLM...\n")

	// Display results
	log.Printf("\n%s\n", strings.Repeat("=", 60))
	log.Printf("ANALYSIS RESULTS\n")
	log.Printf("%s\n", strings.Repeat("=", 60))
	if analysis.JokePercentage == nil {
		log.Printf("Joke Percentage: null (no mention of jokes)\n")
	} else {
		log.Printf("Joke Percentage: %d\n", *analysis.JokePercentage)
	}
	if analysis.JokeReasoning != nil {
		log.Printf("Joke Reasoning: %s\n", *analysis.JokeReasoning)
	}
	log.Printf("%s\n", strings.Repeat("=", 60))
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
		log.Printf("Error: unknown mode '%s'. Valid modes: joke\n", *mode)
		log.Printf("Usage: %s [flags]\n", os.Args[0])
		flag.PrintDefaults()
		log.Fatalf("")
	}

	// Validate that exactly one of --url or --rss is provided
	urlProvided := *url != ""
	rssProvided := *rss != ""

	if !urlProvided && !rssProvided {
		log.Printf("Error: exactly one of --url or --rss must be provided\n")
		log.Printf("Usage: %s [flags]\n", os.Args[0])
		flag.PrintDefaults()
		log.Fatalf("")
	}

	if urlProvided && rssProvided {
		log.Printf("Error: cannot specify both --url and --rss\n")
		log.Printf("Usage: %s [flags]\n", os.Args[0])
		flag.PrintDefaults()
		log.Fatalf("")
	}

	// Get API key from flag, embedded secrets, or environment
	apiKeyValue := *apiKey
	if apiKeyValue == "" {
		// Try embedded key from lib/secrets
		apiKeyValue = lib.OpenAIKey()
		if apiKeyValue == "" {
			// Fall back to environment variable
			apiKeyValue = os.Getenv("OPENAI_API_KEY")
		}
	}

	// Set up Datastore client with embedded credentials
	ctx := context.Background()
	datastoreClient, err := lib.CreateDatastoreClient(ctx)
	if err != nil {
		log.Fatalf("Error creating Datastore client: %v\n", err)
	}
	defer datastoreClient.Close()

	if urlProvided {
		// Single URL mode - existing behavior
		log.Printf("Fetching article from: %s\n", *url)
		page, cachePath, err := fetcher.FetchArticleContent(ctx, *url, *verbose, datastoreClient)
		if err != nil {
			log.Fatalf("Error: %v\n", err)
		}
		_ = cachePath // cache path available for future use

		analysis, err := analyzer.Analyze(ctx, page, apiKeyValue, promptMode, datastoreClient, *verbose)
		if err != nil {
			log.Fatalf("Error: %v\n", err)
		}
		displayAnalysis(analysis, page.Title, page.URL, page.Content, *verbose, 0, 0)
	} else {
		// RSS mode - fetch articles and analyze each
		pages, err := rssfetcher.FetchRSSArticles(ctx, *rss, *max, *verbose, datastoreClient)
		if err != nil {
			log.Fatalf("Error fetching RSS articles: %v\n", err)
		}

		if len(pages) == 0 {
			log.Fatalf("Error: no articles fetched from RSS feed\n")
		}

		log.Printf("\n%s\n", strings.Repeat("=", 60))
		log.Printf("Analyzing %d article(s) from RSS feed\n", len(pages))
		log.Printf("%s\n\n", strings.Repeat("=", 60))

		for i, page := range pages {
			showSeparator := i < len(pages)-1
			analysis, err := analyzer.Analyze(ctx, page, apiKeyValue, promptMode, datastoreClient, *verbose)
			if err != nil {
				log.Printf("Error analyzing article %d: %v\n", i+1, err)
				log.Printf("%s\n", strings.Repeat("-", 120))
				if showSeparator {
					log.Printf("\n")
				}
				continue
			}
			displayAnalysis(analysis, page.Title, page.URL, page.Content, *verbose, i+1, len(pages))
		}
	}
}
