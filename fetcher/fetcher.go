package fetcher

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/PuerkitoBio/goquery"
	"github.com/zeace/poisson/models"
)

const (
	cacheDir        = "cache"
	cacheExpiration = 24 * time.Hour
)

// urlToCacheFilename converts a URL to a safe cache filename using SHA256 hash.
func urlToCacheFilename(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])
}

// getCachePath returns the full path to the cache file for a given URL.
func getFileCachePath(url string) (string, error) {
	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("error creating cache directory: %w", err)
	}

	filename := urlToCacheFilename(url)
	return filepath.Join(cacheDir, filename), nil
}

// saveToFileCache saves content to the cache file.
func saveToFileCache(cachePath, content string) error {
	if err := os.WriteFile(cachePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("error writing cache file: %w", err)
	}
	return nil
}

// FetchArticleContent fetches and extracts text content from a given URL.
// It checks Datastore first, and uses cached content if available.
// If verbose is true, it prints whether it's using cached content or fetching from the URL.
// It will save new pages to Datastore, and into a local file cache.
// Returns a CrawledPage, cache file path, and an error.
func FetchArticleContent(
	ctx context.Context,
	url string,
	verbose bool,
	datastoreClient *datastore.Client,
) (*models.CrawledPage, string, error) {
	var page *models.CrawledPage

	// Get cache path (used in all return cases)
	cachePath, err := getFileCachePath(url)
	if err != nil {
		return nil, "", fmt.Errorf("error getting cache path: %w", err)
	}

	// Check Datastore first
	page, found, err := models.GetCrawledPage(ctx, datastoreClient, url)
	if err != nil {
		return nil, "", fmt.Errorf("error getting crawled page from Datastore: %w", err)
	}
	if found {
		if verbose {
			fmt.Println("Using cached version from Datastore")
		}
		// Ensure content is also in file cache
		if err := saveToFileCache(cachePath, page.Content); err != nil {
			// Log error but don't fail the request
			if verbose {
				fmt.Printf("Warning: failed to save to file cache: %v\n", err)
			}
		}
		return page, cachePath, nil
	}

	// Cache miss, fetch from URL
	if verbose {
		fmt.Println("Fetching from URL...")
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("error fetching URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("error parsing HTML: %w", err)
	}

	// Extract title
	title := doc.Find("title").First().Text()
	title = strings.TrimSpace(title)

	// Remove script and style elements
	doc.Find("script, style").Remove()

	// Try to find main content areas
	var text string
	mainContent := doc.Find("main").First()
	if mainContent.Length() == 0 {
		mainContent = doc.Find("article").First()
	}
	if mainContent.Length() == 0 {
		mainContent = doc.Find("div.content").First()
	}

	if mainContent.Length() > 0 {
		text = mainContent.Text()
	} else {
		// Fallback to body text
		text = doc.Find("body").Text()
	}

	// Clean up whitespace
	text = strings.Join(strings.Fields(text), " ")

	if text == "" {
		return nil, cachePath, fmt.Errorf("no content extracted from URL")
	}

	// Save to Datastore if client is provided
	if datastoreClient != nil && ctx != nil {
		crawlTime := time.Now()
		var err error
		page, err = models.CreateCrawledPage(ctx, datastoreClient, url, title, text, crawlTime)
		if err != nil {
			return nil, "", fmt.Errorf("error saving crawled page to Datastore: %w", err)
		}
		if verbose {
			fmt.Println("Saved to Datastore")
		}
	}

	// Save to cache
	if err := saveToFileCache(cachePath, text); err != nil {
		// Log error but don't fail the request
		// In a production system, you might want to log this
		_ = err
	}

	return page, cachePath, nil
}
