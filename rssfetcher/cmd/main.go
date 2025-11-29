package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/zeace/poisson/rssfetcher"
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

	articles, err := rssfetcher.FetchRSSArticles(*url, *max, *verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n%s\n", strings.Repeat("=", 60))
	fmt.Printf("Fetched %d article(s) from RSS feed\n", len(articles))
	fmt.Printf("%s\n\n", strings.Repeat("=", 60))

	for i, article := range articles {
		fmt.Printf("Article %d (%d characters):\n", i+1, len(article))
		fmt.Println(strings.Repeat("-", 60))
		preview := article
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		fmt.Println(preview)
		fmt.Println(strings.Repeat("-", 60))
		if i < len(articles)-1 {
			fmt.Println()
		}
	}
}

