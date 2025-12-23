package graph

import (
	"github.com/zeace/poisson/lib"
)

// Resolver handles GraphQL queries and mutations
type Resolver struct {
	datastoreClient lib.DatastoreClient
}

// NewResolver creates a new resolver instance
func NewResolver(datastoreClient lib.DatastoreClient) *Resolver {
	return &Resolver{
		datastoreClient: datastoreClient,
	}
}
