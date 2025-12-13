package analyzer

import (
	"encoding/json"
	"fmt"

	"github.com/zeace/poisson/models"
)

// testIntermediateResult is used to parse the LLM response for test mode before converting to AnalysisResult.
type testIntermediateResult struct {
	Result string `json:"result"`
}

// ProcessTestResponse processes the JSON response from the LLM for test mode and converts it to AnalysisResult.
func ProcessTestResponse(jsonStr string, fingerprint int) (*models.AnalysisResult, error) {
	var intermediate testIntermediateResult
	if err := json.Unmarshal([]byte(jsonStr), &intermediate); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	// For test mode, we don't analyze joke percentage, so return nil
	result := &models.AnalysisResult{
		Mode:              AnalysisModeTest,
		JokePercentage:    nil,
		PromptFingerprint: fingerprint,
	}

	return result, nil
}
