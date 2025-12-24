package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/zeace/poisson/crawler/analyzer"
	"github.com/zeace/poisson/lib"
	"github.com/zeace/poisson/models"
)

func TestGetFeed_EmptyResult(t *testing.T) {
	ctx := context.Background()
	mockDS := lib.NewMockDatastoreClient()

	oldestDate := time.Now().Add(-24 * time.Hour)
	items, err := GetFeed(ctx, mockDS, 10, oldestDate, "joke")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(items) != 0 {
		t.Errorf("Expected empty feed, got %d items", len(items))
	}
}

func TestGetFeed_NoAnalysisResults(t *testing.T) {
	ctx := context.Background()
	mockDS := lib.NewMockDatastoreClient()

	// Add a crawled page but no analysis result
	now := time.Now()
	_, err := mockDS.WriteCrawledPage(ctx, "https://example.com/article1", "Article 1", "Content 1", now)
	if err != nil {
		t.Fatalf("Failed to write crawled page: %v", err)
	}

	oldestDate := now.Add(-1 * time.Hour)
	items, err := GetFeed(ctx, mockDS, 10, oldestDate, "joke")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(items) != 0 {
		t.Errorf("Expected empty feed (no analysis results), got %d items", len(items))
	}
}

func TestGetFeed_SingleItem(t *testing.T) {
	ctx := context.Background()
	mockDS := lib.NewMockDatastoreClient()

	// Add a crawled page with analysis result
	now := time.Now()
	page, err := mockDS.WriteCrawledPage(ctx, "https://example.com/article1", "Article 1", "Content 1", now)
	if err != nil {
		t.Fatalf("Failed to write crawled page: %v", err)
	}

	// Add analysis result
	jokePercentage := 75
	result := &models.AnalysisResult{
		Mode:           analyzer.AnalysisModeJoke,
		JokePercentage: &jokePercentage,
	}
	err = mockDS.WriteAnalysisResult(ctx, page.URL, result)
	if err != nil {
		t.Fatalf("Failed to write analysis result: %v", err)
	}

	oldestDate := now.Add(-1 * time.Hour)
	items, err := GetFeed(ctx, mockDS, 10, oldestDate, "joke")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items))
	}

	if items[0].URL != page.URL {
		t.Errorf("Expected URL %s, got %s", page.URL, items[0].URL)
	}

	if items[0].Title != page.Title {
		t.Errorf("Expected title %s, got %s", page.Title, items[0].Title)
	}

	if items[0].JokeConfidence != 75 {
		t.Errorf("Expected joke confidence 75, got %d", items[0].JokeConfidence)
	}
}

func TestGetFeed_MultipleItemsSorted(t *testing.T) {
	ctx := context.Background()
	mockDS := lib.NewMockDatastoreClient()

	now := time.Now()

	// Add multiple crawled pages with different joke percentages
	pages := []struct {
		url         string
		title       string
		jokePercent int
		datetime    time.Time
	}{
		{"https://example.com/article1", "Article 1", 50, now},
		{"https://example.com/article2", "Article 2", 90, now},
		{"https://example.com/article3", "Article 3", 30, now},
		{"https://example.com/article4", "Article 4", 80, now},
	}

	for _, p := range pages {
		page, err := mockDS.WriteCrawledPage(ctx, p.url, p.title, "Content", p.datetime)
		if err != nil {
			t.Fatalf("Failed to write crawled page: %v", err)
		}

		result := &models.AnalysisResult{
			Mode:           analyzer.AnalysisModeJoke,
			JokePercentage: &p.jokePercent,
		}
		err = mockDS.WriteAnalysisResult(ctx, page.URL, result)
		if err != nil {
			t.Fatalf("Failed to write analysis result: %v", err)
		}
	}

	oldestDate := now.Add(-1 * time.Hour)
	items, err := GetFeed(ctx, mockDS, 10, oldestDate, "joke")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(items) != 4 {
		t.Fatalf("Expected 4 items, got %d", len(items))
	}

	// Verify items are sorted by joke confidence (descending)
	expectedOrder := []int{90, 80, 50, 30}
	for i, item := range items {
		if item.JokeConfidence != expectedOrder[i] {
			t.Errorf("Item %d: expected joke confidence %d, got %d", i, expectedOrder[i], item.JokeConfidence)
		}
	}
}

func TestGetFeed_MaxArticlesLimit(t *testing.T) {
	ctx := context.Background()
	mockDS := lib.NewMockDatastoreClient()

	now := time.Now()

	// Add 5 crawled pages
	for i := 1; i <= 5; i++ {
		url := fmt.Sprintf("https://example.com/article%d", i)
		title := fmt.Sprintf("Article %d", i)
		jokePercent := 10 * i

		page, err := mockDS.WriteCrawledPage(ctx, url, title, "Content", now)
		if err != nil {
			t.Fatalf("Failed to write crawled page: %v", err)
		}

		result := &models.AnalysisResult{
			Mode:           analyzer.AnalysisModeJoke,
			JokePercentage: &jokePercent,
		}
		err = mockDS.WriteAnalysisResult(ctx, page.URL, result)
		if err != nil {
			t.Fatalf("Failed to write analysis result: %v", err)
		}
	}

	oldestDate := now.Add(-1 * time.Hour)
	items, err := GetFeed(ctx, mockDS, 3, oldestDate, "joke")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(items) != 3 {
		t.Fatalf("Expected 3 items (maxArticles limit), got %d", len(items))
	}

	// Verify we got the top 3 (highest joke confidence)
	expectedConfidences := []int{50, 40, 30}
	for i, item := range items {
		if item.JokeConfidence != expectedConfidences[i] {
			t.Errorf("Item %d: expected joke confidence %d, got %d", i, expectedConfidences[i], item.JokeConfidence)
		}
	}
}

func TestGetFeed_FiltersByDate(t *testing.T) {
	ctx := context.Background()
	mockDS := lib.NewMockDatastoreClient()

	now := time.Now()
	oldDate := now.Add(-48 * time.Hour) // 2 days ago
	newDate := now.Add(-1 * time.Hour)  // 1 hour ago

	// Add old page
	oldPage, err := mockDS.WriteCrawledPage(ctx, "https://example.com/old", "Old Article", "Content", oldDate)
	if err != nil {
		t.Fatalf("Failed to write old crawled page: %v", err)
	}
	oldPercent := 60
	oldResult := &models.AnalysisResult{
		Mode:           analyzer.AnalysisModeJoke,
		JokePercentage: &oldPercent,
	}
	err = mockDS.WriteAnalysisResult(ctx, oldPage.URL, oldResult)
	if err != nil {
		t.Fatalf("Failed to write old analysis result: %v", err)
	}

	// Add new page
	newPage, err := mockDS.WriteCrawledPage(ctx, "https://example.com/new", "New Article", "Content", newDate)
	if err != nil {
		t.Fatalf("Failed to write new crawled page: %v", err)
	}
	newPercent := 70
	newResult := &models.AnalysisResult{
		Mode:           analyzer.AnalysisModeJoke,
		JokePercentage: &newPercent,
	}
	err = mockDS.WriteAnalysisResult(ctx, newPage.URL, newResult)
	if err != nil {
		t.Fatalf("Failed to write new analysis result: %v", err)
	}

	// Query with oldestDate that should only include the new page
	oldestDate := now.Add(-24 * time.Hour) // 1 day ago
	items, err := GetFeed(ctx, mockDS, 10, oldestDate, "joke")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("Expected 1 item (new page only), got %d", len(items))
	}

	if items[0].URL != newPage.URL {
		t.Errorf("Expected new page URL %s, got %s", newPage.URL, items[0].URL)
	}
}

func TestGetFeed_SkipsPagesWithoutJokePercentage(t *testing.T) {
	ctx := context.Background()
	mockDS := lib.NewMockDatastoreClient()

	now := time.Now()

	// Add page with analysis but no joke percentage (nil)
	page1, err := mockDS.WriteCrawledPage(ctx, "https://example.com/article1", "Article 1", "Content", now)
	if err != nil {
		t.Fatalf("Failed to write crawled page: %v", err)
	}
	result1 := &models.AnalysisResult{
		Mode:           analyzer.AnalysisModeTest, // Test mode doesn't have joke percentage
		JokePercentage: nil,
	}
	err = mockDS.WriteAnalysisResult(ctx, page1.URL, result1)
	if err != nil {
		t.Fatalf("Failed to write analysis result: %v", err)
	}

	// Add page with joke percentage
	page2, err := mockDS.WriteCrawledPage(ctx, "https://example.com/article2", "Article 2", "Content", now)
	if err != nil {
		t.Fatalf("Failed to write crawled page: %v", err)
	}
	jokePercent := 80
	result2 := &models.AnalysisResult{
		Mode:           analyzer.AnalysisModeJoke,
		JokePercentage: &jokePercent,
	}
	err = mockDS.WriteAnalysisResult(ctx, page2.URL, result2)
	if err != nil {
		t.Fatalf("Failed to write analysis result: %v", err)
	}

	oldestDate := now.Add(-1 * time.Hour)
	items, err := GetFeed(ctx, mockDS, 10, oldestDate, "joke")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("Expected 1 item (only page with joke percentage), got %d", len(items))
	}

	if items[0].URL != page2.URL {
		t.Errorf("Expected page2 URL %s, got %s", page2.URL, items[0].URL)
	}
}

func TestGetFeed_InvalidMode(t *testing.T) {
	ctx := context.Background()
	mockDS := lib.NewMockDatastoreClient()

	oldestDate := time.Now().Add(-24 * time.Hour)
	_, err := GetFeed(ctx, mockDS, 10, oldestDate, "invalid-mode")

	if err == nil {
		t.Fatal("Expected error for invalid mode, got nil")
	}
}
