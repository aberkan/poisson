package models

import (
	"cloud.google.com/go/datastore"

	"github.com/zeace/poisson/lib"
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
}

// MakeAnalysisResultKey returns a Datastore key for an AnalysisResult using the URL and mode as the key name.
// The key name format is "url:mode" to ensure uniqueness per URL and mode combination.
// The URL is normalized before creating the key.
func MakeAnalysisResultKey(url, mode string) *datastore.Key {
	normalizedURL := lib.NormalizeURL(url)
	keyName := normalizedURL + ":" + mode
	return datastore.NameKey(AnalysisResultKind, keyName, nil)
}
