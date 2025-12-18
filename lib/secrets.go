package lib

import (
	"embed"
	"strings"
)

//go:embed secrets/openai_key secrets/poisson-berkan-ace77ca9cd3c.json
var secretsFS embed.FS

// OpenAIKey returns the embedded OpenAI API key, trimmed of whitespace
func OpenAIKey() string {
	data, err := secretsFS.ReadFile("secrets/openai_key")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// GoogleKeyJSON returns the embedded Google Cloud service account JSON key
func GoogleKeyJSON() []byte {
	data, err := secretsFS.ReadFile("secrets/poisson-berkan-ace77ca9cd3c.json")
	if err != nil {
		return nil
	}
	return data
}
