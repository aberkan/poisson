package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/zeace/poisson/crawler/analyzer"
	"github.com/zeace/poisson/crawler/config"
	"github.com/zeace/poisson/crawler/fetcher"
	"github.com/zeace/poisson/crawler/rssfetcher"
	"github.com/zeace/poisson/crawler/utils"
	"github.com/zeace/poisson/lib"
	"github.com/zeace/poisson/models"
)

// Config holds the application configuration parsed from command-line flags
type Config struct {
	APIKey  string
	Verbose bool
	URL     string
	RSS     string
	Max     int
	Mode    string
}

func main() {
	cfg := parseFlags()
	validateConfig(cfg)

	apiKey := config.GetOpenAIKey(cfg.APIKey)
	datastoreClient := setupDatastore()
	defer datastoreClient.Close()

	if cfg.URL != "" {
		runURLMode(cfg, apiKey, datastoreClient)
	} else {
		runRSSMode(cfg, apiKey, datastoreClient)
	}
}

// parseFlags parses command-line flags and returns a Config struct
func parseFlags() *Config {
	var (
		apiKey  = flag.String("api-key", "", "OpenAI API key (or set OPENAI_API_KEY environment variable)")
		verbose = flag.Bool("verbose", false, "Show verbose output")
		url     = flag.String("url", "", "URL of the article to analyze")
		rss     = flag.String("rss", "", "URL of the RSS feed to analyze")
		max     = flag.Int("max", 5, "Maximum number of articles to fetch from RSS feed")
		mode    = flag.String("mode", "joke", "Analysis mode (joke)")
	)
	flag.Parse()

	return &Config{
		APIKey:  *apiKey,
		Verbose: *verbose,
		URL:     *url,
		RSS:     *rss,
		Max:     *max,
		Mode:    *mode,
	}
}

// validateConfig validates the configuration and exits with error message if invalid
func validateConfig(cfg *Config) {
	// Validate mode
	_, err := analyzer.VerifyValidMode(cfg.Mode)
	if err != nil {
		log.Printf("Error: unknown mode '%s'. Valid modes: joke\n", cfg.Mode)
		log.Printf("Usage: %s [flags]\n", os.Args[0])
		flag.PrintDefaults()
		log.Fatalf("")
	}

	// Validate that exactly one of --url or --rss is provided
	urlProvided := cfg.URL != ""
	rssProvided := cfg.RSS != ""

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

	// Validate URLs
	if urlProvided {
		if err := utils.ValidateURL(cfg.URL); err != nil {
			log.Fatalf("Invalid URL: %v\n", err)
		}
	} else {
		if err := utils.ValidateRSSURL(cfg.RSS); err != nil {
			log.Fatalf("Invalid RSS feed URL: %v\n", err)
		}
	}
}

// setupDatastore creates and returns a Datastore client
func setupDatastore() lib.DatastoreClient {
	ctx, cancel := config.NewDatastoreContext()
	defer cancel()

	datastoreClient, err := lib.CreateDatastoreClient(ctx)
	if err != nil {
		log.Fatalf("Error creating Datastore client: %v\n", err)
	}
	return datastoreClient
}

// runURLMode handles single URL analysis mode
func runURLMode(cfg *Config, apiKey string, datastoreClient lib.DatastoreClient) {
	promptMode, _ := analyzer.VerifyValidMode(cfg.Mode) // Already validated in validateConfig

	// Fetch article with timeout
	fetchCtx, fetchCancel := config.NewFetchContext()
	defer fetchCancel()

	log.Printf("Fetching article from: %s\n", cfg.URL)
	page, cachePath, err := fetcher.FetchArticleContent(fetchCtx, cfg.URL, cfg.Verbose, datastoreClient)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
	_ = cachePath // cache path available for future use

	// Analyze with timeout
	analysisCtx, analysisCancel := config.NewAnalysisContext()
	defer analysisCancel()

	analysis, err := analyzer.Analyze(analysisCtx, page, apiKey, promptMode, datastoreClient, cfg.Verbose)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
	displayAnalysis(analysis, page.Title, page.URL, page.Content, cfg.Verbose, 0, 0)
}

// runRSSMode handles RSS feed analysis mode
func runRSSMode(cfg *Config, apiKey string, datastoreClient lib.DatastoreClient) {
	promptMode, _ := analyzer.VerifyValidMode(cfg.Mode) // Already validated in validateConfig

	// Fetch articles from RSS feed with timeout
	rssCtx, rssCancel := config.NewRSSContext()
	defer rssCancel()

	pages, err := rssfetcher.FetchRSSArticles(rssCtx, cfg.RSS, cfg.Max, cfg.Verbose, datastoreClient)
	if err != nil {
		// Check if we got partial success (some pages but also errors)
		if len(pages) == 0 {
			// Complete failure - no pages fetched
			log.Fatalf("Error fetching RSS articles: %v\n", err)
		}
		// Partial success - log warning but continue with available pages
		log.Printf("Warning: %v\n", err)
	}

	if len(pages) == 0 {
		log.Fatalf("Error: no articles fetched from RSS feed\n")
	}

	log.Printf("\n%s\n", strings.Repeat("=", 60))
	log.Printf("Analyzing %d article(s) from RSS feed\n", len(pages))
	log.Printf("%s\n\n", strings.Repeat("=", 60))

	for i, page := range pages {
		showSeparator := i < len(pages)-1

		// Analyze each article with timeout
		analysisCtx, analysisCancel := config.NewAnalysisContext()
		analysis, err := analyzer.Analyze(analysisCtx, page, apiKey, promptMode, datastoreClient, cfg.Verbose)
		analysisCancel() // Cancel immediately after analysis to free resources

		if err != nil {
			log.Printf("Error analyzing article %d: %v\n", i+1, err)
			log.Printf("%s\n", strings.Repeat("-", 120))
			if showSeparator {
				log.Printf("\n")
			}
			continue
		}
		displayAnalysis(analysis, page.Title, page.URL, page.Content, cfg.Verbose, i+1, len(pages))
	}
}

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
