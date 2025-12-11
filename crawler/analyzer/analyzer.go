package analyzer

import (
	"context"
	"fmt"

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

// Analyze analyzes content with LLM and returns the analysis result.
func Analyze(page *models.CrawledPage, apiKey string, mode PromptMode) (string, error) {
	prompt, err := GeneratePrompt(mode, page.Title, page.Content)
	if err != nil {
		return "", fmt.Errorf("error generating prompt: %w", err)
	}
	analysis, err := AnalyzeWithLLM(prompt, apiKey)
	if err != nil {
		return "", fmt.Errorf("error analyzing content: %w", err)
	}
	return analysis, nil
}
