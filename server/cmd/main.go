package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/datastore"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/zeace/poisson/lib"
	"github.com/zeace/poisson/server/graph"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()

	// Get project ID from environment or use default
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = "poisson-berkan" // Default from existing code
	}

	// Initialize Datastore client with embedded credentials
	googleKeyJSON := lib.GoogleKeyJSON()
	var datastoreClient *datastore.Client
	var err error
	if len(googleKeyJSON) > 0 {
		// Use embedded credentials
		datastoreClient, err = datastore.NewClient(ctx, projectID, option.WithCredentialsJSON(googleKeyJSON))
	} else {
		// Fall back to default credentials (e.g., from environment)
		datastoreClient, err = datastore.NewClient(ctx, projectID)
	}
	if err != nil {
		log.Printf("Failed to create datastore client: %v", err)
		os.Exit(1)
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

	// GraphQL endpoint with CORS wrapper
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Handle CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		fmt.Printf("graphql request\n")
		graphqlHandler.ServeHTTP(w, r)
	})

	// GraphQL endpoint with CORS wrapper
	mux.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		// Handle CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		fmt.Printf("graphql request\n")
		graphqlHandler.ServeHTTP(w, r)
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/graphiql", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("playground request\n")
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
