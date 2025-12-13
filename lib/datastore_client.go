package lib

import (
	"context"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/zeace/poisson/models"
)

// DatastoreClient defines the interface for all Datastore operations needed by the crawler.
type DatastoreClient interface {
	// CrawledPage operations
	GetCrawledPage(ctx context.Context, url string) (*models.CrawledPage, bool, error)
	CreateCrawledPage(ctx context.Context, url, title, content string, datetime time.Time) (*models.CrawledPage, error)

	// AnalysisResult operations
	GetAnalysisResult(ctx context.Context, url, mode string) (*models.AnalysisResult, bool, error)
	CreateAnalysisResult(ctx context.Context, url string, result *models.AnalysisResult) error
}

// datastoreClientAdapter wraps a *datastore.Client to implement DatastoreClient
type datastoreClientAdapter struct {
	client *datastore.Client
}

// NewDatastoreClient creates a new DatastoreClient from a datastore.Client
func NewDatastoreClient(client *datastore.Client) DatastoreClient {
	return &datastoreClientAdapter{client: client}
}

func (d *datastoreClientAdapter) GetCrawledPage(ctx context.Context, url string) (*models.CrawledPage, bool, error) {
	return models.GetCrawledPage(ctx, d.client, url)
}

func (d *datastoreClientAdapter) CreateCrawledPage(ctx context.Context, url, title, content string, datetime time.Time) (*models.CrawledPage, error) {
	return models.CreateCrawledPage(ctx, d.client, url, title, content, datetime)
}

func (d *datastoreClientAdapter) GetAnalysisResult(ctx context.Context, url, mode string) (*models.AnalysisResult, bool, error) {
	return models.GetAnalysisResult(ctx, d.client, url, mode)
}

func (d *datastoreClientAdapter) CreateAnalysisResult(ctx context.Context, url string, result *models.AnalysisResult) error {
	return models.CreateAnalysisResult(ctx, d.client, url, result)
}

// MockDatastoreClient is a mock implementation of DatastoreClient for testing
type MockDatastoreClient struct {
	Pages            map[string]*models.CrawledPage
	AnalysisResults  map[string]*models.AnalysisResult
	GetError         error
	CreateError      error
	GetAnalysisError error
	CreateAnalysisError error
}

// NewMockDatastoreClient creates a new MockDatastoreClient
func NewMockDatastoreClient() *MockDatastoreClient {
	return &MockDatastoreClient{
		Pages:           make(map[string]*models.CrawledPage),
		AnalysisResults: make(map[string]*models.AnalysisResult),
	}
}

func (m *MockDatastoreClient) GetCrawledPage(ctx context.Context, url string) (*models.CrawledPage, bool, error) {
	if m.GetError != nil {
		return nil, false, m.GetError
	}
	if page, exists := m.Pages[url]; exists {
		return page, true, nil
	}
	return nil, false, nil
}

func (m *MockDatastoreClient) CreateCrawledPage(ctx context.Context, url, title, content string, datetime time.Time) (*models.CrawledPage, error) {
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

func (m *MockDatastoreClient) GetAnalysisResult(ctx context.Context, url, mode string) (*models.AnalysisResult, bool, error) {
	if m.GetAnalysisError != nil {
		return nil, false, m.GetAnalysisError
	}
	key := url + ":" + mode
	if result, exists := m.AnalysisResults[key]; exists {
		return result, true, nil
	}
	return nil, false, nil
}

func (m *MockDatastoreClient) CreateAnalysisResult(ctx context.Context, url string, result *models.AnalysisResult) error {
	if m.CreateAnalysisError != nil {
		return m.CreateAnalysisError
	}
	key := url + ":" + string(result.Mode)
	m.AnalysisResults[key] = result
	return nil
}

