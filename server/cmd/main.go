package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/datastore"
	"github.com/zeace/poisson/server"
)

func main() {
	ctx := context.Background()

	// Get project ID from environment or use default
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = "poisson-berkan" // Default from existing code
	}

	// Initialize Datastore client
	datastoreClient, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create datastore client: %v", err)
	}
	defer datastoreClient.Close()

	// Create GraphQL handler
	graphqlHandler, err := server.NewGraphQLHandler(datastoreClient)
	if err != nil {
		log.Fatalf("Failed to create GraphQL handler: %v", err)
	}

	// Set up HTTP routes
	mux := http.NewServeMux()

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

	// Root path also serves GraphQL
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Handle CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.URL.Path == "/" {
			fmt.Printf("graphql request\n")
			graphqlHandler.ServeHTTP(w, r)
			return
		}

		http.NotFound(w, r)
	})

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
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
