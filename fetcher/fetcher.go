package fetcher

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// FetchArticleContent fetches and extracts text content from a given URL.
func FetchArticleContent(url string) (string, error) {
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

	return text, nil
}

