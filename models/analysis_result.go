package models

import (
	"context"
	"strings"

	"cloud.google.com/go/datastore"
)

// AnalysisResultKind is the Datastore kind name for AnalysisResult entities
const AnalysisResultKind = "AnalysisResult"

type AnalysisMode string

// AnalysisResult represents the parsed JSON result from the LLM analysis.
// It can be stored in Datastore.
type AnalysisResult struct {
	// Mode is the analysis mode used (e.g., "joke", "test").
	Mode AnalysisMode `json:"mode" datastore:"mode"`
	// JokePercentage is the confidence level that the content is a joke.
	// Nil if jokiness was not analyzed.
	JokePercentage *int `json:"joke_percentage" datastore:"joke_percentage"`
	// JokeReasoning is the reasoning provided by the LLM for the joke analysis.
	// Nil if reasoning was not provided.
	JokeReasoning *string `json:"joke_reasoning" datastore:"joke_reasoning"`
	// PromptFingerprint is a uint64 fingerprint of the prompt template used for this analysis.
	PromptFingerprint uint64 `json:"prompt_fingerprint" datastore:"prompt_fingerprint"`
}

// normalizeURL normalizes a URL by removing the protocol (http:// or https://) and query parameters.
// This is a local copy to avoid import cycles with lib.
func normalizeURL(url string) string {
	// Remove http:// or https:// from the front
	normalized := strings.TrimPrefix(url, "https://")
	normalized = strings.TrimPrefix(normalized, "http://")

	// Remove query parameters (everything after ?)
	if idx := strings.Index(normalized, "?"); idx != -1 {
		normalized = normalized[:idx]
	}

	return normalized
}

// MakeAnalysisResultKey returns a Datastore key for an AnalysisResult using the URL and mode as the key name.
// The key name format is "url:mode" to ensure uniqueness per URL and mode combination.
// The URL is normalized before creating the key.
func MakeAnalysisResultKey(url, mode string) *datastore.Key {
	normalizedURL := normalizeURL(url)
	keyName := normalizedURL + ":" + mode
	return datastore.NameKey(AnalysisResultKind, keyName, nil)
}

// ReadAnalysisResult retrieves an AnalysisResult from Datastore by URL and mode.
// Returns the AnalysisResult and true if found, or nil and false if not found.
func ReadAnalysisResult(
	ctx context.Context,
	client *datastore.Client,
	url, mode string,
) (*AnalysisResult, bool, error) {
	var result AnalysisResult
	key := MakeAnalysisResultKey(url, mode)

	err := client.Get(ctx, key, &result)
	if err == datastore.ErrNoSuchEntity {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	return &result, true, nil
}

// WriteAnalysisResult creates and saves a new AnalysisResult to Datastore.
func WriteAnalysisResult(
	ctx context.Context,
	client *datastore.Client,
	url string,
	result *AnalysisResult,
) error {
	key := MakeAnalysisResultKey(url, string(result.Mode))

	_, err := client.Put(ctx, key, result)
	if err != nil {
		return err
	}

	return nil
}
