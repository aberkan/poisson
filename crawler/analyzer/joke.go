package analyzer

import (
	"encoding/json"
	"fmt"

	"github.com/zeace/poisson/models"
)

// jokeIntermediateResult is used to parse the LLM response for joke mode before converting to AnalysisResult.
type jokeIntermediateResult struct {
	IsJoke     bool   `json:"is_joke"`
	Confidence int    `json:"confidence"`
	Reasoning  string `json:"reasoning"`
}

// ProcessJokeResponse processes the JSON response from the LLM for joke mode and converts it to AnalysisResult.
func ProcessJokeResponse(jsonStr string) (*models.AnalysisResult, error) {
	var intermediate jokeIntermediateResult
	if err := json.Unmarshal([]byte(jsonStr), &intermediate); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	// Ensure confidence is between 0 and 100
	confidence := intermediate.Confidence
	// If it's not a joke, invert the confidence
	if !intermediate.IsJoke {
		confidence = 100 - intermediate.Confidence
	}

	if confidence < 0 {
		confidence = 0
	} else if confidence > 100 {
		confidence = 100
	}

	// Convert to AnalysisResult
	result := &models.AnalysisResult{
		Mode:           AnalysisModeJoke,
		JokePercentage: &confidence,
		JokeReasoning:  &intermediate.Reasoning,
	}

	return result, nil
}
