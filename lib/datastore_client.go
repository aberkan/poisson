package lib

import (
	"context"
	"os"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/zeace/poisson/models"
	"google.golang.org/api/option"
)

// DatastoreClient defines the interface for all Datastore operations needed by the crawler.
type DatastoreClient interface {
	// CrawledPage operations
	ReadCrawledPage(ctx context.Context, url string) (*models.CrawledPage, bool, error)
	WriteCrawledPage(ctx context.Context, url, title, content string, datetime time.Time) (*models.CrawledPage, error)

	// AnalysisResult operations
	ReadAnalysisResult(ctx context.Context, url, mode string) (*models.AnalysisResult, bool, error)
	WriteAnalysisResult(ctx context.Context, url string, result *models.AnalysisResult) error

	// Close closes the underlying datastore client
	Close() error
}

// datastoreClientAdapter wraps a *datastore.Client to implement DatastoreClient
type datastoreClientAdapter struct {
	client *datastore.Client
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
	var client *datastore.Client
	var err error
	if len(googleKeyJSON) > 0 {
		client, err = datastore.NewClient(ctx, projectID, option.WithCredentialsJSON(googleKeyJSON))
	} else {
		// Fall back to default credentials (e.g., from environment)
		client, err = datastore.NewClient(ctx, projectID)
	}
	if err != nil {
		return nil, err
	}

	return NewDatastoreClient(client), nil
}

// NewDatastoreClient creates a new DatastoreClient from a datastore.Client
func NewDatastoreClient(client *datastore.Client) DatastoreClient {
	return &datastoreClientAdapter{client: client}
}

func (d *datastoreClientAdapter) ReadCrawledPage(ctx context.Context, url string) (*models.CrawledPage, bool, error) {
	return models.ReadCrawledPage(ctx, d.client, url)
}

func (d *datastoreClientAdapter) WriteCrawledPage(ctx context.Context, url, title, content string, datetime time.Time) (*models.CrawledPage, error) {
	return models.WriteCrawledPage(ctx, d.client, url, title, content, datetime)
}

func (d *datastoreClientAdapter) ReadAnalysisResult(ctx context.Context, url, mode string) (*models.AnalysisResult, bool, error) {
	return models.ReadAnalysisResult(ctx, d.client, url, mode)
}

func (d *datastoreClientAdapter) WriteAnalysisResult(ctx context.Context, url string, result *models.AnalysisResult) error {
	return models.WriteAnalysisResult(ctx, d.client, url, result)
}

func (d *datastoreClientAdapter) Close() error {
	return d.client.Close()
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
		URL:     url,
		Title:   title,
		Content: content,
	}
	m.Pages[url] = page
	return page, nil
}

func (m *MockDatastoreClient) ReadAnalysisResult(ctx context.Context, url, mode string) (*models.AnalysisResult, bool, error) {
	if m.GetAnalysisError != nil {
		return nil, false, m.GetAnalysisError
	}
	key := url + ":" + mode
	if result, exists := m.AnalysisResults[key]; exists {
		return result, true, nil
	}
	return nil, false, nil
}

func (m *MockDatastoreClient) WriteAnalysisResult(ctx context.Context, url string, result *models.AnalysisResult) error {
	if m.CreateAnalysisError != nil {
		return m.CreateAnalysisError
	}
	key := url + ":" + string(result.Mode)
	m.AnalysisResults[key] = result
	return nil
}

func (m *MockDatastoreClient) Close() error {
	return nil
}
