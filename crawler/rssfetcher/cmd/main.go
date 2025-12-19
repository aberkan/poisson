package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/zeace/poisson/crawler/rssfetcher"
	"github.com/zeace/poisson/lib"
)

func main() {
	var (
		verbose = flag.Bool("verbose", false, "Show verbose output")
		max     = flag.Int("max", 5, "Maximum number of articles to fetch")
		url     = flag.String("url", "", "URL of the RSS feed")
	)
	flag.Parse()

	if *url == "" {
		fmt.Fprintf(os.Stderr, "Error: RSS feed URL required\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Set up Datastore client
	ctx := context.Background()
	datastoreClient, err := lib.CreateDatastoreClient(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Datastore client: %v\n", err)
		os.Exit(1)
	}
	defer datastoreClient.Close()

	pages, err := rssfetcher.FetchRSSArticles(ctx, *url, *max, *verbose, datastoreClient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Printf("Fetched %d article(s) from RSS feed\n", len(pages))
	fmt.Printf("%s\n\n", strings.Repeat("=", 60))

	for i, page := range pages {
		fmt.Printf("Article %d: %s\n", i+1, page.URL)
		fmt.Printf("  Title: %s\n", page.Title)
		fmt.Printf("  Crawled at: %s\n", page.DateTime.Format(time.RFC3339))
		fmt.Printf("  Content length: %d characters\n", len(page.Content))
		fmt.Println(strings.Repeat("-", 60))
		preview := page.Content
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		fmt.Println(preview)
		fmt.Println(strings.Repeat("-", 60))
		if i < len(pages)-1 {
			fmt.Println()
		}
	}
}
