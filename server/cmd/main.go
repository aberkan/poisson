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

	// Set up and start the server
	server := setupServer(datastoreClient)
	port := getPort()

	log.Printf("Starting GraphQL server on port %s", port)
	if err := http.ListenAndServe(":"+port, server); err != nil {
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

// setupServer creates and configures the HTTP server with all routes
func setupServer(datastoreClient lib.DatastoreClient) http.Handler {
	// Create GraphQL handler
	graphqlHandler, err := NewGraphQLHandler(datastoreClient)
	if err != nil {
		log.Fatalf("Failed to create GraphQL handler: %v", err)
	}

	playgroundHandler := NewPlaygroundHandler()

	// Set up HTTP routes
	mux := http.NewServeMux()
	setupRoutes(mux, graphqlHandler, playgroundHandler)

	return mux
}

// setupRoutes registers all HTTP routes with the provided mux
func setupRoutes(mux *http.ServeMux, graphqlHandler *handler.Server, playgroundHandler http.Handler) {
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
	mux.HandleFunc("/health", healthHandler)

	// GraphQL playground endpoint
	mux.HandleFunc("/graphiql", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("playground request\n")
		playgroundHandler.ServeHTTP(w, r)
	})
}

// healthHandler handles the /health endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		log.Printf("Failed to encode health response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// getPort returns the server port from environment variable or default
func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
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
