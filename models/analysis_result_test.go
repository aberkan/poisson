package models

import (
	"testing"
)

func TestMakeAnalysisResultKey_NormalizesURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		mode     string
		expected string
	}{
		{
			name:     "normalizes http:// URL",
			url:      "http://example.com/article?key=value",
			mode:     "joke",
			expected: "example.com/article:joke",
		},
		{
			name:     "normalizes https:// URL",
			url:      "https://example.com/article?key=value",
			mode:     "test",
			expected: "example.com/article:test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := MakeAnalysisResultKey(tt.url, tt.mode)
			if key.Name != tt.expected {
				t.Errorf("MakeAnalysisResultKey(%q, %q) key name = %q, want %q", tt.url, tt.mode, key.Name, tt.expected)
			}
		})
	}
}
