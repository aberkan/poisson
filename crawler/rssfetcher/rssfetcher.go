package rssfetcher

import (
	"context"
	"fmt"
	"log"

	"github.com/mmcdole/gofeed"
	"github.com/zeace/poisson/crawler/fetcher"
	"github.com/zeace/poisson/lib"
	"github.com/zeace/poisson/models"
)

// FetchRSSArticles fetches an RSS feed from the given URL and then fetches
// the content of the first maxArticles articles using FetchArticleContent.
// If datastoreClient and ctx are provided, crawled pages will be saved to Datastore.
// Returns a slice of CrawledPage and any errors encountered.
func FetchRSSArticles(ctx context.Context, feedURL string, maxArticles int, verbose bool, datastoreClient lib.DatastoreClient) ([]*models.CrawledPage, error) {
	if verbose {
		log.Printf("Fetching RSS feed from: %s\n", feedURL)
	}

	// Parse the RSS feed
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(feedURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing RSS feed: %w", err)
	}

	if verbose {
		log.Printf("Found %d items in RSS feed\n", len(feed.Items))
	}

	// Limit to maxArticles
	itemsToFetch := maxArticles
	if len(feed.Items) < itemsToFetch {
		itemsToFetch = len(feed.Items)
	}

	if verbose {
		log.Printf("Fetching first %d articles...\n", itemsToFetch)
	}

	var pages []*models.CrawledPage
	var fetchErrors []error

	for i := 0; i < itemsToFetch; i++ {
		item := feed.Items[i]
		articleURL := item.Link

		if articleURL == "" {
			if verbose {
				log.Printf("Skipping item %d: no URL found\n", i+1)
			}
			continue
		}

		if verbose {
			log.Printf("\n[%d/%d] Fetching: %s\n", i+1, itemsToFetch, articleURL)
			if item.Title != "" {
				log.Printf("  Title: %s\n", item.Title)
			}
		}

		page, _, err := fetcher.FetchArticleContent(ctx, articleURL, verbose, datastoreClient)
		if err != nil {
			fetchErrors = append(fetchErrors, fmt.Errorf("article %s: %w", articleURL, err))
			if verbose {
				log.Printf("  Error: %v\n", err)
			}
			continue
		}

		pages = append(pages, page)
	}

	// If we have errors and no pages, return an error
	if len(pages) == 0 && len(fetchErrors) > 0 {
		return nil, fmt.Errorf("failed to fetch any articles: %v", fetchErrors)
	}

	// If we got some pages but also some errors, return pages with an error indicating partial failure
	if len(pages) > 0 && len(fetchErrors) > 0 {
		if verbose {
			log.Printf("\nWarning: %d article(s) fetched successfully, but %d error(s) occurred:\n", len(pages), len(fetchErrors))
			for _, err := range fetchErrors {
				log.Printf("  - %v\n", err)
			}
		}
		return pages, fmt.Errorf("partial success: fetched %d article(s) but %d error(s) occurred: %v",
			len(pages), len(fetchErrors), fetchErrors)
	}

	return pages, nil
}
