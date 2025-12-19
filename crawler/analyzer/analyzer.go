package analyzer

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/zeace/poisson/lib"
	"github.com/zeace/poisson/models"
)

// AnalyzeWithLLM analyzes content using an LLM with the provided prompt.
// Deprecated: Use LlmClient interface instead. This function is kept for backward compatibility.
func AnalyzeWithLLM(prompt, apiKey string) (string, error) {
	ctx := context.Background()
	client := NewGptLlmClient(apiKey)
	return client.Analyze(ctx, prompt)
}

// parseJSONResponse extracts and parses JSON from the LLM response.
// It handles cases where the response might be wrapped in markdown code blocks or have extra text.
func parseJSONResponse(response string) (string, error) {
	// Remove markdown code blocks if present
	response = strings.TrimSpace(response)
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimSuffix(response, "```")
	} else if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
	}
	response = strings.TrimSpace(response)

	// Try to find JSON object boundaries
	startIdx := strings.Index(response, "{")
	endIdx := strings.LastIndex(response, "}")
	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		return "", fmt.Errorf("no JSON object found in response")
	}

	return response[startIdx : endIdx+1], nil
}

// analyze is the internal function that analyzes content with LLM and returns the parsed analysis result.
// It uses the lib.DatastoreClient interface directly.
func analyze(
	ctx context.Context,
	page *models.CrawledPage,
	llmClient LlmClient,
	mode AnalysisMode,
	datastoreClient lib.DatastoreClient,
	verbose bool,
) (*models.AnalysisResult, error) {
	// Generate prompt fingerprint for this mode
	fingerprint, err := GeneratePromptFingerprint(mode)
	if err != nil {
		return nil, fmt.Errorf("error generating prompt fingerprint: %w", err)
	}

	// Check cache in datastore
	cachedResult, found, err := datastoreClient.ReadAnalysisResult(ctx, page.URL, string(mode))
	if err != nil {
		return nil, fmt.Errorf("error checking analysis cache: %w", err)
	}
	if found {
		// Verify that the PromptFingerprint matches before using cached result
		if cachedResult.PromptFingerprint == fingerprint {
			if verbose {
				log.Printf("Using cached analysis result from Datastore\n")
			}
			return cachedResult, nil
		}
		// Fingerprint doesn't match, continue to analyze with LLM
		if verbose {
			log.Printf("Cached result has mismatched fingerprint, analyzing with LLM...\n")
		}
	}

	// Cache miss or fingerprint mismatch, analyze with LLM
	if verbose {
		log.Printf("Analyzing with LLM...\n")
	}
	prompt, err := GeneratePrompt(mode, page.Title, page.Content)
	if err != nil {
		return nil, fmt.Errorf("error generating prompt: %w", err)
	}
	rawResponse, err := llmClient.Analyze(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("error analyzing content: %w", err)
	}

	// Parse JSON from response
	jsonStr, err := parseJSONResponse(rawResponse)
	if err != nil {
		return nil, fmt.Errorf("error extracting JSON from response: %w", err)
	}

	// Get processing function from prompt config
	config, ok := PromptTemplates[mode]
	if !ok {
		return nil, fmt.Errorf("unknown mode: %s", mode)
	}

	// Process response using the mode-specific processing function
	result, err := config.ProcessResponse(jsonStr, fingerprint)
	if err != nil {
		return nil, err
	}

	// Save to cache
	err = datastoreClient.WriteAnalysisResult(ctx, page.URL, result)
	if err != nil {
		log.Printf("Warning: error saving analysis result to cache: %v\n", err)
		// The analysis was successful, caching is just an optimization
	} else if verbose {
		log.Printf("Saved analysis result to Datastore cache\n")
	}

	return result, nil
}

// Analyze analyzes content with LLM and returns the parsed analysis result.
// If datastoreClient is provided, it will check for cached results and save new results.
func Analyze(
	ctx context.Context,
	page *models.CrawledPage,
	apiKey string,
	mode AnalysisMode,
	datastoreClient lib.DatastoreClient,
	verbose bool,
) (*models.AnalysisResult, error) {
	llmClient := NewGptLlmClient(apiKey)
	return analyze(ctx, page, llmClient, mode, datastoreClient, verbose)
}
