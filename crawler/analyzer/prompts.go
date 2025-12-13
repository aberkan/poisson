package analyzer

import (
	_ "embed"
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/zeace/poisson/models"
)

const maxContentLength = 8000

type AnalysisMode = models.AnalysisMode

const (
	AnalysisModeJoke AnalysisMode = "joke"
	AnalysisModeTest AnalysisMode = "test"
)

//go:embed prompts/joke.prompt.md
var JokePromptTemplate string

//go:embed prompts/test.prompt.md
var TestPromptTemplate string

// PromptConfig holds the template and processing function for a prompt mode.
type PromptConfig struct {
	Template        string
	ProcessResponse func(string, int) (*models.AnalysisResult, error)
}

var PromptTemplates = map[AnalysisMode]PromptConfig{
	AnalysisModeJoke: {
		Template:        JokePromptTemplate,
		ProcessResponse: ProcessJokeResponse,
	},
	AnalysisModeTest: {
		Template:        TestPromptTemplate,
		ProcessResponse: ProcessTestResponse,
	},
}

// VerifyValidMode checks if the given mode is valid (exists in PromptTemplates).
func VerifyValidMode(mode string) (AnalysisMode, error) {
	analysisMode := AnalysisMode(strings.ToLower(mode))
	_, ok := PromptTemplates[analysisMode]
	if !ok {
		return "", fmt.Errorf("unknown mode '%s'", mode)
	}
	return analysisMode, nil
}

// AddBodyToPrompt merges the title and body content into the prompt template.
func AddBodyToPrompt(template, title, body string) string {
	return fmt.Sprintf(template, title, body)
}

// GeneratePromptFingerprint generates an int fingerprint based on the template text for a given mode.
func GeneratePromptFingerprint(mode AnalysisMode) (int, error) {
	config, ok := PromptTemplates[mode]
	if !ok {
		return 0, fmt.Errorf("unknown mode '%s'", mode)
	}

	// Use FNV-1a hash for 64-bit fingerprint, then convert to int
	h := fnv.New64a()
	h.Write([]byte(config.Template))
	return int(h.Sum64()), nil
}

// GeneratePrompt generates a prompt by selecting the appropriate template based on mode
// and merging it with the provided title and content. Content is truncated if it exceeds maxContentLength.
func GeneratePrompt(mode AnalysisMode, title, content string) (string, error) {
	config, ok := PromptTemplates[mode]
	if !ok {
		return "", fmt.Errorf("unknown mode '%s'", mode)
	}

	// Truncate content if too long
	truncatedContent := content
	if len(truncatedContent) > maxContentLength {
		truncatedContent = truncatedContent[:maxContentLength] + "... [content truncated]"
	}

	return AddBodyToPrompt(config.Template, title, truncatedContent), nil
}
