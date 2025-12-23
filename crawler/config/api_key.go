package config

import (
	"os"

	"github.com/zeace/poisson/lib"
)

// GetOpenAIKey returns the OpenAI API key from the following sources in order:
// 1. flagValue (if provided)
// 2. Embedded key from lib/secrets
// 3. OPENAI_API_KEY environment variable
func GetOpenAIKey(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}

	// Try embedded key from lib/secrets
	if key := lib.OpenAIKey(); key != "" {
		return key
	}

	// Fall back to environment variable
	return os.Getenv("OPENAI_API_KEY")
}

