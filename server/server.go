package server

import (
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/zeace/poisson/server/graph"
)

// NewGraphQLHandler creates a new GraphQL handler using gqlgen
func NewGraphQLHandler(datastoreClient *datastore.Client) (*handler.Server, error) {
	// Create resolver
	resolver := graph.NewResolver(datastoreClient)

	// Create executable schema
	executableSchema := graph.NewExecutableSchema(graph.Config{
		Resolvers: resolver,
	})

	// Create GraphQL handler
	srv := handler.NewDefaultServer(executableSchema)

	return srv, nil
}

// NewPlaygroundHandler creates a GraphQL playground handler
func NewPlaygroundHandler() http.Handler {
	return playground.Handler("GraphQL playground", "/graphql")
}
