package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/zeace/poisson/fetcher"
)

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Error: URL argument required\n")
		fmt.Fprintf(os.Stderr, "Usage: %s <url>\n", os.Args[0])
		os.Exit(1)
	}

	url := flag.Arg(0)

	fmt.Printf("Fetching article from: %s\n", url)
	content, err := fetcher.FetchArticleContent(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nFetched %d characters of content\n\n", len(content))
	fmt.Println("Content:")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println(content)
	fmt.Println(strings.Repeat("=", 60))
}

