package fetcher

import (
	"context"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/zeace/poisson/models"
)

// DatastoreClient defines the interface for Datastore operations needed by the fetcher.
type DatastoreClient interface {
	GetCrawledPage(ctx context.Context, url string) (*models.CrawledPage, bool, error)
	CreateCrawledPage(ctx context.Context, url, title, content string, datetime time.Time) (*models.CrawledPage, error)
}

// datastoreClientAdapter wraps a *datastore.Client to implement DatastoreClient
type datastoreClientAdapter struct {
	client *datastore.Client
}

func (d *datastoreClientAdapter) GetCrawledPage(ctx context.Context, url string) (*models.CrawledPage, bool, error) {
	return models.GetCrawledPage(ctx, d.client, url)
}

func (d *datastoreClientAdapter) CreateCrawledPage(ctx context.Context, url, title, content string, datetime time.Time) (*models.CrawledPage, error) {
	return models.CreateCrawledPage(ctx, d.client, url, title, content, datetime)
}

// MockDatastoreClient is a mock implementation of DatastoreClient for testing
type MockDatastoreClient struct {
	Pages       map[string]*models.CrawledPage
	GetError    error
	CreateError error
}

func NewMockDatastoreClient() *MockDatastoreClient {
	return &MockDatastoreClient{
		Pages: make(map[string]*models.CrawledPage),
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
