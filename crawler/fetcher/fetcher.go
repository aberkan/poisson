package fetcher

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/zeace/poisson/lib"
	"github.com/zeace/poisson/models"
)

const (
	cacheDir = "cache"
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

// fetchArticleContent is an internal function that fetches and extracts text content from a given URL.
// It checks Datastore first, and uses cached content if available.
// If verbose is true, it prints whether it's using cached content or fetching from the URL.
// It will save new pages to Datastore, and into the provided cache writer.
// httpClient is used for making HTTP requests.
// cacheWriter is used for writing content to the file cache.
// datastoreClient can be nil, in which case Datastore operations will be skipped.
// normalizedURL is the normalized URL (without protocol and query params) used for Datastore operations.
// Returns a CrawledPage, cache file path, and an error.
func fetchArticleContent(
	ctx context.Context,
	normalizedURL string,
	verbose bool,
	datastoreClient lib.DatastoreClient,
	httpClient *http.Client,
	cacheWriter io.Writer,
	cachePath string,
) (*models.CrawledPage, string, error) {
	var page *models.CrawledPage

	// Check Datastore first using normalized URL
	var found bool
	var err error
	page, found, err = datastoreClient.ReadCrawledPage(ctx, normalizedURL)
	if err != nil {
		return nil, "", fmt.Errorf("error getting crawled page from Datastore: %w", err)
	}
	if found {
		if verbose {
			log.Printf("Using cached version from Datastore\n")
		}
		// Ensure content is also in file cache
		if _, err := cacheWriter.Write([]byte(page.Content)); err != nil {
			// Log error but don't fail the request
			if verbose {
				log.Printf("Warning: failed to save to file cache: %v\n", err)
			}
		}
		return page, cachePath, nil
	}

	// Cache miss, fetch from URL
	// Add protocol back for HTTP request
	fetchURL := lib.AddProtocol(normalizedURL)
	if verbose {
		log.Printf("Fetching from URL...\n")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fetchURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := httpClient.Do(req)
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

	// Save to Datastore using normalized URL
	crawlTime := time.Now()
	page, err = datastoreClient.WriteCrawledPage(ctx, normalizedURL, title, text, crawlTime)
	if err != nil {
		return nil, "", fmt.Errorf("error saving crawled page to Datastore: %w", err)
	}
	if verbose {
		log.Printf("Saved to Datastore\n")
	}

	// Save to cache
	if _, err := cacheWriter.Write([]byte(text)); err != nil {
		// Log error but don't fail the request
		// In a production system, you might want to log this
		_ = err
	}

	return page, cachePath, nil
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
	datastoreClient lib.DatastoreClient,
) (*models.CrawledPage, string, error) {
	// Normalize URL for Datastore operations (remove protocol and query params)
	normalizedURL := lib.NormalizeURL(url)

	// Get cache path (used in all return cases) - use normalized URL for cache
	cachePath, err := getFileCachePath(normalizedURL)
	if err != nil {
		return nil, "", fmt.Errorf("error getting cache path: %w", err)
	}

	// Create HTTP client
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Open cache file for writing
	cacheFile, err := os.OpenFile(cachePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, "", fmt.Errorf("error opening cache file: %w", err)
	}
	defer cacheFile.Close()

	// Use normalized URL for all operations
	return fetchArticleContent(ctx, normalizedURL, verbose, datastoreClient, httpClient, cacheFile, cachePath)
}
