package analyzer

import (
	"context"
	"fmt"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// AddBodyToPrompt merges the body content into the prompt template.
func AddBodyToPrompt(template, body string) string {
	return fmt.Sprintf(template, body)
}

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
