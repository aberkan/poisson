package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/zeace/poisson/models"
)

// AnalyzeWithLLM analyzes content using an LLM with the provided prompt.
func AnalyzeWithLLM(prompt, apiKey string) (string, error) {

	client := openai.NewClient(option.WithAPIKey(apiKey))
	ctx := context.Background()

	chatCompletion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model: openai.ChatModelGPT4o,
	})

	if err != nil {
		return "", fmt.Errorf("error calling OpenAI API: %w", err)
	}

	if len(chatCompletion.Choices) == 0 {
		return "", fmt.Errorf("no choices in OpenAI response: %v", chatCompletion)
	}

	return chatCompletion.Choices[0].Message.Content, nil
}

// AnalysisResult represents the parsed JSON result from the LLM analysis.
type AnalysisResult struct {
	// JokePercentage is the confidence level that the content is a joke.
	// Nil if jokiness was not analyzed.
	JokePercentage *int `json:"joke_percentage"`
}

// intermediateResult is used to parse the LLM response before converting to AnalysisResult.
type intermediateResult struct {
	IsJoke     bool   `json:"is_joke"`
	Confidence int    `json:"confidence"`
	Reasoning  string `json:"reasoning"`
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

// Analyze analyzes content with LLM and returns the parsed analysis result.
func Analyze(page *models.CrawledPage, apiKey string, mode PromptMode) (*AnalysisResult, error) {
	prompt, err := GeneratePrompt(mode, page.Title, page.Content)
	if err != nil {
		return nil, fmt.Errorf("error generating prompt: %w", err)
	}
	rawResponse, err := AnalyzeWithLLM(prompt, apiKey)
	if err != nil {
		return nil, fmt.Errorf("error analyzing content: %w", err)
	}

	// Parse JSON from response
	jsonStr, err := parseJSONResponse(rawResponse)
	if err != nil {
		return nil, fmt.Errorf("error extracting JSON from response: %w", err)
	}

	var intermediate intermediateResult
	if err := json.Unmarshal([]byte(jsonStr), &intermediate); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	// Convert to AnalysisResult
	result := &AnalysisResult{
		JokePercentage: &intermediate.Confidence,
	}

	return result, nil
}
