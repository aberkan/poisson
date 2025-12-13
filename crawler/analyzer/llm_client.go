package analyzer

import (
	"context"
	"fmt"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// LlmClient defines the interface for LLM operations.
type LlmClient interface {
	// Analyze analyzes content using an LLM with the provided prompt.
	Analyze(ctx context.Context, prompt string) (string, error)
}

// GptLlmClient is an implementation of LlmClient that uses OpenAI's GPT API.
type GptLlmClient struct {
	apiKey string
}

// NewGptLlmClient creates a new GptLlmClient with the provided API key.
func NewGptLlmClient(apiKey string) *GptLlmClient {
	return &GptLlmClient{apiKey: apiKey}
}

// Analyze analyzes content using OpenAI's GPT API.
func (g *GptLlmClient) Analyze(ctx context.Context, prompt string) (string, error) {
	client := openai.NewClient(option.WithAPIKey(g.apiKey))

	chatCompletion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model: openai.ChatModelGPT4o,
	})

	if err != nil {
		return "", err
	}

	if len(chatCompletion.Choices) == 0 {
		return "", fmt.Errorf("no choices in OpenAI response")
	}

	return chatCompletion.Choices[0].Message.Content, nil
}

// MockLlmClient is a mock implementation of LlmClient for testing.
type MockLlmClient struct {
	Response string
	Error    error
}

// Analyze returns the mock response or error.
func (m *MockLlmClient) Analyze(ctx context.Context, prompt string) (string, error) {
	if m.Error != nil {
		return "", m.Error
	}
	return m.Response, nil
}
