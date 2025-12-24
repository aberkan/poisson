package server

import (
	"context"
	"sort"
	"time"

	"github.com/zeace/poisson/crawler/analyzer"
	"github.com/zeace/poisson/lib"
)

// FeedItem represents a single item in the feed
type FeedItem struct {
	URL            string
	Title          string
	JokeConfidence int // JokePercentage from AnalysisResult
}

// GetFeed retrieves analysis results since oldest_date, ranks them by jokeConfidence,
// and returns up to max_articles items.
// It uses the CrawledPage DateTime to filter by date since AnalysisResult doesn't have a timestamp.
func GetFeed(
	ctx context.Context,
	datastoreClient lib.DatastoreClient,
	maxArticles int,
	oldestDate time.Time,
	modeStr string,
) ([]FeedItem, error) {
	// Get all CrawledPages since oldestDate
	pages, err := datastoreClient.GetCrawledPagesSince(ctx, oldestDate)
	if err != nil {
		return nil, err
	}
	mode, err := analyzer.VerifyValidMode(modeStr)
	if err != nil {
		return nil, err
	}

	// For each page, get its analysis result and build feed items
	var items []FeedItem

	for _, page := range pages {
		// Try to get analysis result for the specified mode
		analysis, found, err := datastoreClient.ReadAnalysisResult(ctx, page.URL, mode)
		if err != nil {
			continue // Skip on error
		}
		if !found || analysis.JokePercentage == nil {
			continue // Skip if no analysis or no joke percentage
		}

		items = append(items, FeedItem{
			URL:            page.URL,
			Title:          page.Title,
			JokeConfidence: *analysis.JokePercentage,
		})
	}

	// Sort by joke confidence (descending)
	sort.Slice(items, func(i, j int) bool {
		return items[i].JokeConfidence > items[j].JokeConfidence
	})

	// Take up to maxArticles
	if len(items) > maxArticles {
		items = items[:maxArticles]
	}

	return items, nil
}
