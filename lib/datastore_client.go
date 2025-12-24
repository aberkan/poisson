package lib

import (
	"context"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/zeace/poisson/models"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DatastoreClient defines the interface for all Datastore operations needed by the crawler.
type DatastoreClient interface {
	// CrawledPage operations
	ReadCrawledPage(ctx context.Context, url string) (*models.CrawledPage, bool, error)
	WriteCrawledPage(ctx context.Context, url, title, content string, datetime time.Time) (*models.CrawledPage, error)
	GetCrawledPagesSince(ctx context.Context, oldestDate time.Time) ([]models.CrawledPage, error)

	// AnalysisResult operations
	ReadAnalysisResult(ctx context.Context, url string, mode models.AnalysisMode) (*models.AnalysisResult, bool, error)
	WriteAnalysisResult(ctx context.Context, url string, result *models.AnalysisResult) error

	// Close closes the underlying datastore client
	Close() error
}

// datastoreClientAdapter wraps a *firestore.Client to implement DatastoreClient
type datastoreClientAdapter struct {
	client *firestore.Client
}

// CreateDatastoreClient creates a new DatastoreClient with embedded credentials or default credentials.
// It uses the project ID from GOOGLE_CLOUD_PROJECT environment variable, or defaults to "poisson-berkan".
func CreateDatastoreClient(ctx context.Context) (DatastoreClient, error) {
	// Get project ID from environment or use default
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = "poisson-berkan"
	}

	// Try to use embedded credentials first
	googleKeyJSON := GoogleKeyJSON()
	var client *firestore.Client
	var err error
	if len(googleKeyJSON) > 0 {
		client, err = firestore.NewClient(ctx, projectID, option.WithCredentialsJSON(googleKeyJSON))
	} else {
		// Fall back to default credentials (e.g., from environment)
		client, err = firestore.NewClient(ctx, projectID)
	}
	if err != nil {
		return nil, err
	}

	return NewDatastoreClient(client), nil
}

// NewDatastoreClient creates a new DatastoreClient from a firestore.Client
func NewDatastoreClient(client *firestore.Client) DatastoreClient {
	return &datastoreClientAdapter{client: client}
}

func (d *datastoreClientAdapter) ReadCrawledPage(ctx context.Context, url string) (*models.CrawledPage, bool, error) {
	key := UrlToCrawledPageKey(url)
	docRef := d.client.Collection(models.CrawledPageKind).Doc(key)
	doc, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, false, nil
		}
		return nil, false, err
	}

	var page models.CrawledPage
	if err := doc.DataTo(&page); err != nil {
		return nil, false, err
	}
	page.URL = url // Ensure URL is set from original URL (not the key)

	return &page, true, nil
}

func (d *datastoreClientAdapter) WriteCrawledPage(ctx context.Context, url, title, content string, datetime time.Time) (*models.CrawledPage, error) {
	if datetime.IsZero() {
		datetime = time.Now()
	}

	page := &models.CrawledPage{
		URL:      url,
		Title:    title,
		Content:  content,
		DateTime: datetime,
	}

	key := UrlToCrawledPageKey(url)
	docRef := d.client.Collection(models.CrawledPageKind).Doc(key)
	_, err := docRef.Set(ctx, page)
	if err != nil {
		return nil, err
	}

	return page, nil
}

// GetCrawledPagesSince returns all CrawledPages with DateTime >= oldestDate.
func (d *datastoreClientAdapter) GetCrawledPagesSince(ctx context.Context, oldestDate time.Time) ([]models.CrawledPage, error) {
	query := d.client.Collection(models.CrawledPageKind).
		Where("DateTime", ">=", oldestDate).OrderBy("DateTime", firestore.Desc)

	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	var pages []models.CrawledPage
	for _, doc := range docs {
		var page models.CrawledPage
		if err := doc.DataTo(&page); err != nil {
			continue // Skip invalid documents
		}
		// URL is already set from document data when we wrote it
		pages = append(pages, page)
	}

	return pages, nil
}

func (d *datastoreClientAdapter) ReadAnalysisResult(
	ctx context.Context,
	url string,
	mode models.AnalysisMode,
) (*models.AnalysisResult, bool, error) {
	// Convert URL to analysis key
	keyName := UrlToAnalysisKey(url, mode)

	docRef := d.client.Collection(models.AnalysisResultKind).Doc(keyName)
	doc, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, false, nil
		}
		return nil, false, err
	}

	var result models.AnalysisResult
	if err := doc.DataTo(&result); err != nil {
		return nil, false, err
	}

	return &result, true, nil
}

func (d *datastoreClientAdapter) WriteAnalysisResult(ctx context.Context, url string, result *models.AnalysisResult) error {
	// Convert URL to analysis key
	keyName := UrlToAnalysisKey(url, result.Mode)

	docRef := d.client.Collection(models.AnalysisResultKind).Doc(keyName)
	_, err := docRef.Set(ctx, result)
	if err != nil {
		return err
	}

	return nil
}

func (d *datastoreClientAdapter) Close() error {
	return d.client.Close()
}

// normalizeURL normalizes a URL by removing the protocol (http:// or https://) and query parameters.
func normalizeURL(url string) string {
	// Remove http:// or https:// from the front
	normalized := strings.TrimPrefix(url, "https://")
	normalized = strings.TrimPrefix(normalized, "http://")

	// Remove query parameters (everything after ?)
	if idx := strings.Index(normalized, "?"); idx != -1 {
		normalized = normalized[:idx]
	}

	return normalized
}

// UrlToCrawledPageKey converts a URL to a key suitable for use as a CrawledPage document ID.
// It removes query parameters, trailing slashes, and converts '/' to '_'.
func UrlToCrawledPageKey(url string) string {
	key := url

	// Remove query parameters (everything after ?)
	if idx := strings.Index(key, "?"); idx != -1 {
		key = key[:idx]
	}

	// Remove trailing slashes
	key = strings.TrimRight(key, "/")

	// Convert '/' to '_'
	key = strings.ReplaceAll(key, "/", "_")

	return key
}

// UrlToAnalysisKey converts a URL to a key suitable for use as an AnalysisResult document ID.
// It removes query parameters, trailing slashes, and converts '/' to '_'.
// This is used as part of the key name along with the mode (format: "key:mode").
func UrlToAnalysisKey(url string, mode models.AnalysisMode) string {
	// Normalize URL (remove protocol)
	key := UrlToCrawledPageKey(url)
	key += ":" + string(mode)

	return key
}

// MockDatastoreClient is a mock implementation of DatastoreClient for testing
type MockDatastoreClient struct {
	Pages               map[string]*models.CrawledPage
	AnalysisResults     map[string]*models.AnalysisResult
	GetError            error
	CreateError         error
	GetAnalysisError    error
	CreateAnalysisError error
}

// NewMockDatastoreClient creates a new MockDatastoreClient
func NewMockDatastoreClient() *MockDatastoreClient {
	return &MockDatastoreClient{
		Pages:           make(map[string]*models.CrawledPage),
		AnalysisResults: make(map[string]*models.AnalysisResult),
	}
}

func (m *MockDatastoreClient) ReadCrawledPage(ctx context.Context, url string) (*models.CrawledPage, bool, error) {
	if m.GetError != nil {
		return nil, false, m.GetError
	}
	if page, exists := m.Pages[url]; exists {
		return page, true, nil
	}
	return nil, false, nil
}

func (m *MockDatastoreClient) WriteCrawledPage(ctx context.Context, url, title, content string, datetime time.Time) (*models.CrawledPage, error) {
	if m.CreateError != nil {
		return nil, m.CreateError
	}
	page := &models.CrawledPage{
		URL:      url,
		Title:    title,
		Content:  content,
		DateTime: datetime,
	}
	m.Pages[url] = page
	return page, nil
}

func (m *MockDatastoreClient) GetCrawledPagesSince(ctx context.Context, oldestDate time.Time) ([]models.CrawledPage, error) {
	var pages []models.CrawledPage

	for _, page := range m.Pages {
		if !page.DateTime.IsZero() && !page.DateTime.Before(oldestDate) {
			pages = append(pages, *page)
		}
	}

	return pages, nil
}

func (m *MockDatastoreClient) ReadAnalysisResult(ctx context.Context, url string, mode models.AnalysisMode) (*models.AnalysisResult, bool, error) {
	if m.GetAnalysisError != nil {
		return nil, false, m.GetAnalysisError
	}
	key := UrlToAnalysisKey(url, mode)
	if result, exists := m.AnalysisResults[key]; exists {
		return result, true, nil
	}
	return nil, false, nil
}

func (m *MockDatastoreClient) WriteAnalysisResult(ctx context.Context, url string, result *models.AnalysisResult) error {
	if m.CreateAnalysisError != nil {
		return m.CreateAnalysisError
	}
	key := UrlToAnalysisKey(url, result.Mode)
	m.AnalysisResults[key] = result
	return nil
}

func (m *MockDatastoreClient) Close() error {
	return nil
}
