package analyzer

import (
	_ "embed"
	"fmt"
	"strings"
)

const maxContentLength = 8000

type PromptMode string

const (
	PromptModeJoke PromptMode = "joke"
	PromptModeTest PromptMode = "test"
)

//go:embed prompts/joke.prompt.md
var JokePromptTemplate string

//go:embed prompts/test.prompt.md
var TestPromptTemplate string

var PromptTemplates = map[PromptMode]string{
	PromptModeJoke: JokePromptTemplate,
	PromptModeTest: TestPromptTemplate,
}

// VerifyValidMode checks if the given mode is valid (exists in PromptTemplates).
func VerifyValidMode(mode string) (PromptMode, error) {
	promptMode := PromptMode(strings.ToLower(mode))
	_, ok := PromptTemplates[promptMode]
	if !ok {
		return "", fmt.Errorf("unknown mode '%s'", mode)
	}
	return promptMode, nil
}

// AddBodyToPrompt merges the title and body content into the prompt template.
func AddBodyToPrompt(template, title, body string) string {
	return fmt.Sprintf(template, title, body)
}

// GeneratePrompt generates a prompt by selecting the appropriate template based on mode
// and merging it with the provided title and content. Content is truncated if it exceeds maxContentLength.
func GeneratePrompt(mode PromptMode, title, content string) (string, error) {
	template, ok := PromptTemplates[mode]
	if !ok {
		return "", fmt.Errorf("unknown mode '%s'", mode)
	}

	// Truncate content if too long
	truncatedContent := content
	if len(truncatedContent) > maxContentLength {
		truncatedContent = truncatedContent[:maxContentLength] + "... [content truncated]"
	}

	return AddBodyToPrompt(template, title, truncatedContent), nil
}
