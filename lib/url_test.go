package lib

import (
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove http://",
			input:    "http://example.com/article",
			expected: "example.com/article",
		},
		{
			name:     "remove https://",
			input:    "https://example.com/article",
			expected: "example.com/article",
		},
		{
			name:     "remove query parameters",
			input:    "example.com/article?key=value&foo=bar",
			expected: "example.com/article",
		},
		{
			name:     "remove http:// and query parameters",
			input:    "http://example.com/article?key=value",
			expected: "example.com/article",
		},
		{
			name:     "remove https:// and query parameters",
			input:    "https://example.com/article?key=value&foo=bar",
			expected: "example.com/article",
		},
		{
			name:     "no protocol or query params",
			input:    "example.com/article",
			expected: "example.com/article",
		},
		{
			name:     "only query params",
			input:    "example.com/article?key=value",
			expected: "example.com/article",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "URL with fragment",
			input:    "https://example.com/article#section",
			expected: "example.com/article#section",
		},
		{
			name:     "URL with query and fragment (fragment removed with query)",
			input:    "https://example.com/article?key=value#section",
			expected: "example.com/article",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeURL(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

