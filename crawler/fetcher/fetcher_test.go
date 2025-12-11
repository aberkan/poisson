package fetcher

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/zeace/poisson/models"
)

func TestFetchArticleContent_FromURL(t *testing.T) {
	const htmlContent = `<!DOCTYPE html>
<html>
<head>
	<title>Test Article</title>
</head>
<body>
	<main>
		<h1>Test Article Content</h1>
		<p>This is the main content of the article.</p>
		<p>It has multiple paragraphs.</p>
	</main>
	<script>console.log("ignore me");</script>
	<style>body { color: red; }</style>
</body>
</html>`

	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		userAgent := r.Header.Get("User-Agent")
		if userAgent != "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" {
			t.Errorf("Expected User-Agent header, got %s", userAgent)
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(htmlContent))
	}))
	defer server.Close()

	ctx := context.Background()
	httpClient := &http.Client{Timeout: 5 * time.Second}
	var cacheWriter bytes.Buffer
	cachePath := "/test/cache/path"
	mockDS := NewMockDatastoreClient()

	page, path, err := fetchArticleContent(ctx, server.URL, false, mockDS, httpClient, &cacheWriter, cachePath)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if page == nil {
		t.Fatal("Expected page to be non-nil, but got nil")
	}

	if page.Title != "Test Article" {
		t.Errorf("Expected title 'Test Article', got '%s'", page.Title)
	}

	if !strings.Contains(page.Content, "Test Article Content") {
		t.Errorf("Expected content to contain 'Test Article Content', got: %s", page.Content)
	}
	// Verify script and style were removed
	if strings.Contains(page.Content, "console.log") || strings.Contains(page.Content, "color: red") {
		t.Error("Expected script and style tags to be removed from content")
	}

	if path != cachePath {
		t.Errorf("Expected cache path '%s', got '%s'", cachePath, path)
	}

	// Verify content was written to cache
	cachedContent := cacheWriter.String()
	if cachedContent != page.Content {
		t.Errorf("Expected cache content to match page content, but they differ. Cache: %s, Page: %s", cachedContent, page.Content)
	}

	// Verify page was saved to mock Datastore
	savedPage, found, _ := mockDS.GetCrawledPage(ctx, server.URL)
	if !found {
		t.Error("Expected page to be saved to Datastore")
	}
	if savedPage == nil || savedPage.Title != "Test Article" {
		t.Errorf("Expected saved page to have title 'Test Article'")
	}
}

func TestFetchArticleContent_FromDatastoreCache(t *testing.T) {
	ctx := context.Background()
	mockDS := NewMockDatastoreClient()

	// Pre-populate the mock Datastore with a cached page
	cachedPage := &models.CrawledPage{
		URL:      "https://example.com/article",
		Title:    "Cached Article",
		Content:  "This is cached content from Datastore",
		DateTime: time.Now(),
	}
	mockDS.Pages["https://example.com/article"] = cachedPage

	httpClient := &http.Client{Timeout: 5 * time.Second}
	var cacheWriter bytes.Buffer
	cachePath := "/test/cache/path"

	page, path, err := fetchArticleContent(ctx, "https://example.com/article", false, mockDS, httpClient, &cacheWriter, cachePath)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if page == nil {
		t.Fatal("Expected page to be non-nil, but got nil")
	}

	if page.Title != "Cached Article" {
		t.Errorf("Expected title 'Cached Article', got '%s'", page.Title)
	}

	if page.Content != "This is cached content from Datastore" {
		t.Errorf("Expected content 'This is cached content from Datastore', got '%s'", page.Content)
	}

	if path != cachePath {
		t.Errorf("Expected cache path '%s', got '%s'", cachePath, path)
	}

	// Verify content was written to file cache
	cachedContent := cacheWriter.String()
	if cachedContent != "This is cached content from Datastore" {
		t.Errorf("Expected cache content to match page content, got: %s", cachedContent)
	}
}

func TestFetchArticleContent_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ctx := context.Background()
	httpClient := &http.Client{Timeout: 5 * time.Second}
	var cacheWriter bytes.Buffer
	cachePath := "/test/cache/path"
	mockDS := NewMockDatastoreClient()

	_, _, err := fetchArticleContent(ctx, server.URL, false, mockDS, httpClient, &cacheWriter, cachePath)

	if err == nil {
		t.Fatal("Expected error for 500 status code, but got nil")
	}

	if !strings.Contains(err.Error(), "unexpected status code: 500") {
		t.Errorf("Expected error message to contain 'unexpected status code: 500', got: %v", err)
	}
}

func TestFetchArticleContent_FallbackToBody(t *testing.T) {
	htmlContent := `<!DOCTYPE html>
<html>
<head>
	<title>No Main Tag</title>
</head>
<body>
	<h1>Body Content</h1>
	<p>This content is in the body tag directly.</p>
</body>
</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(htmlContent))
	}))
	defer server.Close()

	ctx := context.Background()
	httpClient := &http.Client{Timeout: 5 * time.Second}
	var cacheWriter bytes.Buffer
	cachePath := "/test/cache/path"
	mockDS := NewMockDatastoreClient()

	page, _, err := fetchArticleContent(ctx, server.URL, false, mockDS, httpClient, &cacheWriter, cachePath)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if page == nil {
		t.Fatal("Expected page to be non-nil, but got nil")
	}

	if !strings.Contains(page.Content, "Body Content") {
		t.Errorf("Expected content to contain 'Body Content', got: %s", page.Content)
	}
}

func TestFetchArticleContent_ArticleTag(t *testing.T) {
	htmlContent := `<!DOCTYPE html>
<html>
<head>
	<title>Article Tag</title>
</head>
<body>
	<article>
		<h1>Article Content</h1>
		<p>This is inside an article tag.</p>
	</article>
</body>
</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(htmlContent))
	}))
	defer server.Close()

	ctx := context.Background()
	httpClient := &http.Client{Timeout: 5 * time.Second}
	var cacheWriter bytes.Buffer
	cachePath := "/test/cache/path"
	mockDS := NewMockDatastoreClient()

	page, _, err := fetchArticleContent(ctx, server.URL, false, mockDS, httpClient, &cacheWriter, cachePath)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if page == nil {
		t.Fatal("Expected page to be non-nil, but got nil")
	}

	if !strings.Contains(page.Content, "Article Content") {
		t.Errorf("Expected content to contain 'Article Content', got: %s", page.Content)
	}
}

func TestFetchArticleContent_DatastoreGetError(t *testing.T) {
	ctx := context.Background()
	mockDS := NewMockDatastoreClient()
	mockDS.GetError = errors.New("datastore connection error")

	httpClient := &http.Client{Timeout: 5 * time.Second}
	var cacheWriter bytes.Buffer
	cachePath := "/test/cache/path"

	_, _, err := fetchArticleContent(ctx, "https://example.com/article", false, mockDS, httpClient, &cacheWriter, cachePath)

	if err == nil {
		t.Fatal("Expected error from Datastore, but got nil")
	}

	if !strings.Contains(err.Error(), "error getting crawled page from Datastore") {
		t.Errorf("Expected error message to contain 'error getting crawled page from Datastore', got: %v", err)
	}
}

func TestFetchArticleContent_DatastoreCreateError(t *testing.T) {
	htmlContent := `<!DOCTYPE html>
<html>
<head>
	<title>Create Error</title>
</head>
<body>
	<main>
		<p>Content that should fail to save.</p>
	</main>
</body>
</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(htmlContent))
	}))
	defer server.Close()

	ctx := context.Background()
	mockDS := NewMockDatastoreClient()
	mockDS.CreateError = errors.New("datastore save error")

	httpClient := &http.Client{Timeout: 5 * time.Second}
	var cacheWriter bytes.Buffer
	cachePath := "/test/cache/path"

	_, _, err := fetchArticleContent(ctx, server.URL, false, mockDS, httpClient, &cacheWriter, cachePath)

	if err == nil {
		t.Fatal("Expected error from Datastore create, but got nil")
	}

	if !strings.Contains(err.Error(), "error saving crawled page to Datastore") {
		t.Errorf("Expected error message to contain 'error saving crawled page to Datastore', got: %v", err)
	}
}
