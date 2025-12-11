package models

import (
	"context"
	"time"

	"cloud.google.com/go/datastore"
)

// CrawledPageKind is the Datastore kind name for CrawledPage entities
const CrawledPageKind = "CrawledPage"

// CrawledPage represents a crawled web page stored in Datastore
type CrawledPage struct {
	URL      string    `datastore:"url"`
	Title    string    `datastore:"title"`
	Content  string    `datastore:"content,noindex"`
	DateTime time.Time `datastore:"datetime"`
}

// Key returns a Datastore key for a CrawledPage using the URL as the key name
func (cp *CrawledPage) Key(kind string) *datastore.Key {
	return datastore.NameKey(kind, cp.URL, nil)
}

// GetCrawledPage retrieves a CrawledPage from Datastore by URL
// Returns the CrawledPage and true if found, or nil and false if not found
func GetCrawledPage(ctx context.Context, client *datastore.Client, url string) (*CrawledPage, bool, error) {
	page := &CrawledPage{URL: url}
	key := page.Key(CrawledPageKind)

	err := client.Get(ctx, key, page)
	if err == datastore.ErrNoSuchEntity {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	return page, true, nil
}

// CreateCrawledPage creates and saves a new CrawledPage to Datastore
// If datetime is zero, it will be set to the current time
func CreateCrawledPage(
	ctx context.Context,
	client *datastore.Client,
	url, title, content string,
	datetime time.Time) (*CrawledPage, error) {
	if datetime.IsZero() {
		datetime = time.Now()
	}

	page := &CrawledPage{
		URL:      url,
		Title:    title,
		Content:  content,
		DateTime: datetime,
	}

	key := page.Key(CrawledPageKind)

	_, err := client.Put(ctx, key, page)
	if err != nil {
		return nil, err
	}

	return page, nil
}
