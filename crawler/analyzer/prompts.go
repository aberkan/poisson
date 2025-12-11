package analyzer

import (
	_ "embed"
	"fmt"
)

const maxContentLength = 8000

type PromptMode string

const (
	PromptModeJoke PromptMode = "joke"
)

//go:embed prompts/joke.prompt.md
var JokePromptTemplate string

var PromptTemplates = map[PromptMode]string{
	PromptModeJoke: JokePromptTemplate,
}

// VerifyValidMode checks if the given mode is valid (exists in PromptTemplates).
func VerifyValidMode(mode string) (PromptMode, error) {
	_, ok := PromptTemplates[PromptMode(mode)]
	if !ok {
		return "", fmt.Errorf("unknown mode '%s'", mode)
	}
	return PromptMode(mode), nil
}

// AddBodyToPrompt merges the body content into the prompt template.
func AddBodyToPrompt(template, body string) string {
	return fmt.Sprintf(template, body)
}

// GeneratePrompt generates a prompt by selecting the appropriate template based on mode
// and merging it with the provided content. Content is truncated if it exceeds maxContentLength.
func GeneratePrompt(mode PromptMode, content string) (string, error) {
	template, ok := PromptTemplates[mode]
	if !ok {
		return "", fmt.Errorf("unknown mode '%s'", mode)
	}

	// Truncate content if too long
	truncatedContent := content
	if len(truncatedContent) > maxContentLength {
		truncatedContent = truncatedContent[:maxContentLength] + "... [content truncated]"
	}

	return AddBodyToPrompt(template, truncatedContent), nil
}
