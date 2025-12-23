package config

import (
	"context"
	"time"
)

const (
	// DatastoreTimeout is the timeout for creating Datastore clients
	DatastoreTimeout = 10 * time.Second

	// FetchTimeout is the timeout for fetching article content
	FetchTimeout = 30 * time.Second

	// AnalysisTimeout is the timeout for analyzing content with LLM
	AnalysisTimeout = 60 * time.Second

	// RSSTimeout is the timeout for fetching and processing RSS feeds
	RSSTimeout = 5 * time.Minute
)

// NewContextWithTimeout creates a new context with the specified timeout
func NewContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// NewDatastoreContext creates a context with DatastoreTimeout for Datastore operations
func NewDatastoreContext() (context.Context, context.CancelFunc) {
	return NewContextWithTimeout(DatastoreTimeout)
}

// NewFetchContext creates a context with FetchTimeout for fetching operations
func NewFetchContext() (context.Context, context.CancelFunc) {
	return NewContextWithTimeout(FetchTimeout)
}

// NewAnalysisContext creates a context with AnalysisTimeout for analysis operations
func NewAnalysisContext() (context.Context, context.CancelFunc) {
	return NewContextWithTimeout(AnalysisTimeout)
}

// NewRSSContext creates a context with RSSTimeout for RSS feed operations
func NewRSSContext() (context.Context, context.CancelFunc) {
	return NewContextWithTimeout(RSSTimeout)
}
