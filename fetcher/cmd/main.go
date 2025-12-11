package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/zeace/poisson/fetcher"
)

func main() {
	var (
		verbose = flag.Bool("verbose", false, "Show verbose output")
	)
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Error: URL argument required\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <url>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	url := flag.Arg(0)

	// Set up Datastore client
	ctx := context.Background()
	projectID := "poisson-berkan"
	datastoreClient, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Datastore client: %v\n", err)
		os.Exit(1)
	}
	defer datastoreClient.Close()

	fmt.Printf("Fetching article from: %s\n", url)
	page, cachePath, err := fetcher.FetchArticleContent(ctx, url, *verbose, datastoreClient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Title: %s\n", page.Title)
	fmt.Printf("Cache file: %s\n", cachePath)
	fmt.Printf("Crawled at: %s\n", page.DateTime.Format(time.RFC3339))

	fmt.Printf("\nFetched %d characters of content\n\n", len(page.Content))
	fmt.Println("Content:")
	fmt.Println(strings.Repeat("=", 60))
	if len(page.Content) > 1000 {
		fmt.Println(page.Content[:1000] + "...")
	} else {
		fmt.Println(page.Content)
	}
	fmt.Println(strings.Repeat("=", 60))
}
