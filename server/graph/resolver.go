package graph

import (
	"cloud.google.com/go/datastore"
)

// Resolver handles GraphQL queries and mutations
type Resolver struct {
	datastoreClient *datastore.Client
}

// NewResolver creates a new resolver instance
func NewResolver(datastoreClient *datastore.Client) *Resolver {
	return &Resolver{
		datastoreClient: datastoreClient,
	}
}
