package fetcher

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
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
func getCachePath(url string) (string, error) {
	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("error creating cache directory: %w", err)
	}

	filename := urlToCacheFilename(url)
	return filepath.Join(cacheDir, filename), nil
}

// loadFromCache attempts to load content from the cache.
// If the cache file is older than 24 hours, it deletes the file and returns empty string (cache miss).
func loadFromCache(url string) (string, error) {
	cachePath, err := getCachePath(url)
	if err != nil {
		return "", err
	}

	// Check if file exists and get its modification time
	fileInfo, err := os.Stat(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // Cache miss, not an error
		}
		return "", fmt.Errorf("error checking cache file: %w", err)
	}

	// Check if cache file is older than 24 hours
	if time.Since(fileInfo.ModTime()) > cacheExpiration {
		// Delete expired cache file
		if err := os.Remove(cachePath); err != nil {
			return "", fmt.Errorf("error deleting expired cache file: %w", err)
		}
		return "", nil // Cache miss due to expiration
	}

	// Cache is valid, read and return content
	content, err := os.ReadFile(cachePath)
	if err != nil {
		return "", fmt.Errorf("error reading cache file: %w", err)
	}

	return string(content), nil
}

// saveToCache saves content to the cache.
func saveToCache(url, content string) error {
	cachePath, err := getCachePath(url)
	if err != nil {
		return err
	}

	if err := os.WriteFile(cachePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("error writing cache file: %w", err)
	}

	return nil
}

// FetchArticleContent fetches and extracts text content from a given URL.
// It checks the cache first and uses cached content if available.
// If verbose is true, it prints whether it's using cached content or fetching from the URL.
func FetchArticleContent(url string, verbose bool) (string, error) {
	// Check cache first
	cachedContent, err := loadFromCache(url)
	if err != nil {
		return "", err
	}
	if cachedContent != "" {
		if verbose {
			fmt.Println("Using cached version")
		}
		return cachedContent, nil
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
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error fetching URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error parsing HTML: %w", err)
	}

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
		return "", fmt.Errorf("no content extracted from URL")
	}

	// Save to cache
	if err := saveToCache(url, text); err != nil {
		// Log error but don't fail the request
		// In a production system, you might want to log this
		_ = err
	}

	return text, nil
}

