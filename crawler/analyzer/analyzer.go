package analyzer

import (
	"context"
	"fmt"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

)

const (
	maxContentLength = 8000
)

// AnalyzeWithLLM analyzes article content using an LLM to determine if it's an April Fools joke.
func AnalyzeWithLLM(content, apiKey string) (string, error) {
	// Truncate content if too long
	if len(content) > maxContentLength {
		content = content[:maxContentLength] + "... [content truncated]"
	}

	prompt := fmt.Sprintf(`Analyze the following article and determine if it's an April Fools joke or prank.

Consider these factors:
- Unrealistic or absurd claims
- Date of publication (April 1st is a strong indicator)
- Tone and style (humorous, satirical, or intentionally misleading)
- References to April Fools or pranks
- Outlandish but plausible-sounding claims
- Context clues that suggest it's a joke

Article content:
%s

Provide your analysis in the following format:
- Is it a joke? (Yes/No/Uncertain)
- Confidence level (0-100)
- Reasoning (2-3 sentences explaining your assessment)
- Key indicators that led to your conclusion`, content)

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
		return "", fmt.Errorf("no choices in OpenAI response", chatCompletion)
	}

	return chatCompletion.Choices[0].Message.Content, nil
}

