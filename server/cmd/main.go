package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/zeace/poisson/lib"
	"github.com/zeace/poisson/server/graph"
)

func main() {
	ctx := context.Background()

	// Initialize Datastore client with embedded credentials
	datastoreClient, err := lib.CreateDatastoreClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create datastore client: %v", err)
	}
	defer datastoreClient.Close()

	// Create GraphQL handler
	graphqlHandler, err := NewGraphQLHandler(datastoreClient)
	if err != nil {
		log.Fatalf("Failed to create GraphQL handler: %v", err)
	}

	playgroundHandler := NewPlaygroundHandler()

	// Set up HTTP routes
	mux := http.NewServeMux()

	// GraphQL endpoints with CORS middleware
	mux.HandleFunc("/", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("graphql request\n")
		graphqlHandler.ServeHTTP(w, r)
	}))

	mux.HandleFunc("/graphql", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("graphql request\n")
		graphqlHandler.ServeHTTP(w, r)
	}))

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/graphiql", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("playground request\n")
		playgroundHandler.ServeHTTP(w, r)
	})

	// Get port from environment (Cloud Run sets PORT)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting GraphQL server on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func NewGraphQLHandler(datastoreClient lib.DatastoreClient) (*handler.Server, error) {
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

// corsMiddleware wraps an HTTP handler with CORS headers and OPTIONS handling
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight OPTIONS requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next(w, r)
	}
}
