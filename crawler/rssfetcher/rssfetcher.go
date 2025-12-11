package rssfetcher

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/datastore"
	"github.com/mmcdole/gofeed"
	"github.com/zeace/poisson/crawler/fetcher"
	"github.com/zeace/poisson/models"
)

// FetchRSSArticles fetches an RSS feed from the given URL and then fetches
// the content of the first maxArticles articles using FetchArticleContent.
// If datastoreClient and ctx are provided, crawled pages will be saved to Datastore.
// Returns a slice of CrawledPage and any errors encountered.
func FetchRSSArticles(ctx context.Context, feedURL string, maxArticles int, verbose bool, datastoreClient *datastore.Client) ([]*models.CrawledPage, error) {
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

	var pages []*models.CrawledPage
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

		page, _, err := fetcher.FetchArticleContent(ctx, articleURL, verbose, datastoreClient)
		if err != nil {
			errMsg := fmt.Sprintf("error fetching article %s: %v", articleURL, err)
			errors = append(errors, errMsg)
			if verbose {
				fmt.Printf("  Error: %v\n", err)
			}
			continue
		}

		pages = append(pages, page)
	}

	// If we got some pages but also some errors, return what we have
	// but include error information
	if len(pages) > 0 && len(errors) > 0 {
		if verbose {
			fmt.Printf("\nWarning: %d article(s) fetched successfully, but %d error(s) occurred:\n", len(pages), len(errors))
			for _, errMsg := range errors {
				fmt.Printf("  - %s\n", errMsg)
			}
		}
		return pages, nil
	}

	// If we have errors and no pages, return an error
	if len(pages) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("failed to fetch any articles: %s", strings.Join(errors, "; "))
	}

	return pages, nil
}
