package rssfetcher

import (
	"fmt"
	"strings"

	"github.com/mmcdole/gofeed"
	"github.com/zeace/poisson/fetcher"
)

// FetchRSSArticles fetches an RSS feed from the given URL and then fetches
// the content of the first maxArticles articles using FetchArticleContent.
// Returns a slice of article contents and any errors encountered.
func FetchRSSArticles(feedURL string, maxArticles int, verbose bool) ([]string, error) {
	if verbose {
		fmt.Printf("Fetching RSS feed from: %s\n", feedURL)
	}

	// Parse the RSS feed
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing RSS feed: %w", err)
	}

	if verbose {
		fmt.Printf("Found %d items in RSS feed\n", len(feed.Items))
	}

	// Limit to maxArticles
	itemsToFetch := maxArticles
	if len(feed.Items) < itemsToFetch {
		itemsToFetch = len(feed.Items)
	}

	if verbose {
		fmt.Printf("Fetching first %d articles...\n", itemsToFetch)
	}

	var articles []string
	var errors []string

	for i := 0; i < itemsToFetch; i++ {
		item := feed.Items[i]
		articleURL := item.Link

		if articleURL == "" {
			if verbose {
				fmt.Printf("Skipping item %d: no URL found\n", i+1)
			}
			continue
		}

		if verbose {
			fmt.Printf("\n[%d/%d] Fetching: %s\n", i+1, itemsToFetch, articleURL)
			if item.Title != "" {
				fmt.Printf("  Title: %s\n", item.Title)
			}
		}

		content, err := fetcher.FetchArticleContent(articleURL, verbose)
		if err != nil {
			errMsg := fmt.Sprintf("error fetching article %s: %v", articleURL, err)
			errors = append(errors, errMsg)
			if verbose {
				fmt.Printf("  Error: %v\n", err)
			}
			continue
		}

		articles = append(articles, content)
	}

	// If we got some articles but also some errors, return what we have
	// but include error information
	if len(articles) > 0 && len(errors) > 0 {
		if verbose {
			fmt.Printf("\nWarning: %d article(s) fetched successfully, but %d error(s) occurred:\n", len(articles), len(errors))
			for _, errMsg := range errors {
				fmt.Printf("  - %s\n", errMsg)
			}
		}
		return articles, nil
	}

	// If we have errors and no articles, return an error
	if len(articles) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("failed to fetch any articles: %s", strings.Join(errors, "; "))
	}

	return articles, nil
}

