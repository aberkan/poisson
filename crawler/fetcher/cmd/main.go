package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/zeace/poisson/crawler/config"
	"github.com/zeace/poisson/crawler/fetcher"
	"github.com/zeace/poisson/crawler/utils"
	"github.com/zeace/poisson/lib"
)

func main() {
	var (
		verbose = flag.Bool("verbose", false, "Show verbose output")
	)
	flag.Parse()

	if flag.NArg() == 0 {
		log.Printf("Error: URL argument required\n")
		log.Printf("Usage: %s [flags] <url>\n", os.Args[0])
		flag.PrintDefaults()
		log.Fatalf("")
	}

	url := flag.Arg(0)

	// Validate URL before fetching
	if err := utils.ValidateURL(url); err != nil {
		log.Fatalf("Invalid URL: %v\n", err)
	}

	// Set up Datastore client
	dsCtx, dsCancel := config.NewDatastoreContext()
	datastoreClient, err := lib.CreateDatastoreClient(dsCtx)
	dsCancel()
	if err != nil {
		log.Fatalf("Error creating Datastore client: %v\n", err)
	}
	defer datastoreClient.Close()

	// Fetch article with timeout
	fetchCtx, fetchCancel := config.NewFetchContext()
	defer fetchCancel()

	log.Printf("Fetching article from: %s\n", url)
	page, cachePath, err := fetcher.FetchArticleContent(fetchCtx, url, *verbose, datastoreClient)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	log.Printf("Title: %s\n", page.Title)
	log.Printf("Cache file: %s\n", cachePath)
	log.Printf("Crawled at: %s\n", page.DateTime.Format(time.RFC3339))

	log.Printf("\nFetched %d characters of content\n\n", len(page.Content))
	log.Printf("Content:\n")
	log.Printf("%s\n", strings.Repeat("=", 60))
	if len(page.Content) > 1000 {
		log.Printf("%s\n", page.Content[:1000]+"...")
	} else {
		log.Printf("%s\n", page.Content)
	}
	log.Printf("%s\n", strings.Repeat("=", 60))
}
