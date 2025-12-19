package main

import (
	"context"
	"flag"
	"log"
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
		log.Printf("Error: RSS feed URL required\n")
		log.Printf("Usage: %s [flags]\n", os.Args[0])
		flag.PrintDefaults()
		log.Fatalf("")
	}

	// Set up Datastore client
	ctx := context.Background()
	datastoreClient, err := lib.CreateDatastoreClient(ctx)
	if err != nil {
		log.Fatalf("Error creating Datastore client: %v\n", err)
	}
	defer datastoreClient.Close()

	pages, err := rssfetcher.FetchRSSArticles(ctx, *url, *max, *verbose, datastoreClient)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	log.Printf("\n%s\n", strings.Repeat("=", 60))
	log.Printf("Fetched %d article(s) from RSS feed\n", len(pages))
	log.Printf("%s\n\n", strings.Repeat("=", 60))

	for i, page := range pages {
		log.Printf("Article %d: %s\n", i+1, page.URL)
		log.Printf("  Title: %s\n", page.Title)
		log.Printf("  Crawled at: %s\n", page.DateTime.Format(time.RFC3339))
		log.Printf("  Content length: %d characters\n", len(page.Content))
		log.Printf("%s\n", strings.Repeat("-", 60))
		preview := page.Content
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		log.Printf("%s\n", preview)
		log.Printf("%s\n", strings.Repeat("-", 60))
		if i < len(pages)-1 {
			log.Printf("\n")
		}
	}
}
