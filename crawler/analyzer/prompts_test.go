package analyzer

import (
	"strings"
	"testing"
)

func TestVerifyValidMode(t *testing.T) {
	tests := []struct {
		name        string
		mode        string
		expectError bool
	}{
		{
			name:        "valid joke mode",
			mode:        "joke",
			expectError: false,
		},
		{
			name:        "valid test mode",
			mode:        "test",
			expectError: false,
		},
		{
			name:        "invalid mode",
			mode:        "invalid",
			expectError: true,
		},
		{
			name:        "empty mode",
			mode:        "",
			expectError: true,
		},
		{
			name:        "case sensitive",
			mode:        "JOKE",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode, err := VerifyValidMode(tt.mode)
			if tt.expectError {
				if err == nil {
					t.Errorf("VerifyValidMode(%q) expected error, but got nil. Mode: %v", tt.mode, mode)
				}
			} else {
				if err != nil {
					t.Errorf("VerifyValidMode(%q) returned error: %v", tt.mode, err)
				}
				if mode != PromptMode(strings.ToLower(tt.mode)) {
					t.Errorf("VerifyValidMode(%q) = %v, expected %v", tt.mode, mode, PromptMode(tt.mode))
				}
			}
		})
	}
}

func TestGeneratePrompt_ValidModes(t *testing.T) {
	tests := []struct {
		name             string
		mode             PromptMode
		title            string
		content          string
		expectedInPrompt []string
		shouldContain    bool
	}{
		{
			name:             "joke mode with title and content",
			mode:             PromptModeJoke,
			title:            "Test Article Title",
			content:          "This is test content for joke detection.",
			expectedInPrompt: []string{"Test Article Title", "This is test content for joke detection."},
			shouldContain:    true,
		},
		{
			name:             "test mode with title and content",
			mode:             PromptModeTest,
			title:            "Test Title",
			content:          "This is test content for testing.",
			expectedInPrompt: []string{"Test Title", "This is test content for testing.", "Test prompt template"},
			shouldContain:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt, err := GeneratePrompt(tt.mode, tt.title, tt.content)
			if err != nil {
				t.Fatalf("GeneratePrompt(%q, %q, %q) returned error: %v", tt.mode, tt.title, tt.content, err)
			}

			if prompt == "" {
				t.Error("GeneratePrompt returned empty prompt")
			}

			for _, expected := range tt.expectedInPrompt {
				if tt.shouldContain && !strings.Contains(prompt, expected) {
					t.Errorf("Expected prompt to contain %q, but got: %s", expected, prompt)
				}
			}
		})
	}
}

func TestGeneratePrompt_InvalidMode(t *testing.T) {
	tests := []struct {
		name        string
		mode        PromptMode
		content     string
		expectedErr string
	}{
		{
			name:        "invalid mode",
			mode:        PromptMode("invalid"),
			content:     "some content",
			expectedErr: "unknown mode",
		},
		{
			name:        "empty mode",
			mode:        PromptMode(""),
			content:     "some content",
			expectedErr: "unknown mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt, err := GeneratePrompt(tt.mode, "Test Title", tt.content)
			if err == nil {
				t.Errorf("GeneratePrompt(%q, %q) expected error, but got nil. Prompt: %s", tt.mode, tt.content, prompt)
				return
			}

			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("Expected error to contain %q, but got: %v", tt.expectedErr, err)
			}
		})
	}
}

func TestGeneratePrompt_ContentTruncation(t *testing.T) {
	// Create content longer than maxContentLength
	longContent := strings.Repeat("a", maxContentLength+1000)

	prompt, err := GeneratePrompt(PromptModeJoke, "Test Title", longContent)
	if err != nil {
		t.Fatalf("GeneratePrompt returned error: %v", err)
	}

	// Check that truncation marker is present
	if !strings.Contains(prompt, "... [content truncated]") {
		t.Error("Expected prompt to contain truncation marker, but it didn't")
	}

	// Check that the prompt doesn't exceed reasonable length
	// (template + truncated content + marker)
	if len(prompt) > len(JokePromptTemplate)+maxContentLength+100 {
		t.Errorf("Prompt seems too long, length: %d", len(prompt))
	}
}

func TestGeneratePrompt_TestModeFormat(t *testing.T) {
	content := "Test content here"
	prompt, err := GeneratePrompt(PromptModeTest, "Test Title", content)
	if err != nil {
		t.Fatalf("GeneratePrompt returned error: %v", err)
	}

	// Verify the test prompt template format
	if !strings.Contains(prompt, "Test prompt template") {
		t.Error("Expected prompt to contain 'Test prompt template'")
	}

	if !strings.Contains(prompt, content) {
		t.Errorf("Expected prompt to contain content %q, but got: %s", content, prompt)
	}

	// Verify it follows the template format (should have "Content: " followed by content)
	if !strings.Contains(prompt, "Content:") {
		t.Error("Expected prompt to contain 'Content:' from test template")
	}
}

func TestAddBodyToPrompt(t *testing.T) {
	template := "Template with title: %s and content: %s"
	title := "Test Title"
	body := "Test body content"

	result := AddBodyToPrompt(template, title, body)

	if !strings.Contains(result, body) {
		t.Errorf("Expected result to contain body %q, but got: %s", body, result)
	}

	if !strings.Contains(result, title) {
		t.Errorf("Expected result to contain title %q, but got: %s", title, result)
	}

	if !strings.Contains(result, "Template with title:") {
		t.Error("Expected result to contain template text")
	}

	expected := "Template with title: Test Title and content: Test body content"
	if result != expected {
		t.Errorf("Expected %q, but got %q", expected, result)
	}
}

func TestPromptTemplates_AllModesPresent(t *testing.T) {
	expectedModes := []string{"joke", "test"}

	for _, modeStr := range expectedModes {
		mode, err := VerifyValidMode(modeStr)
		if err != nil {
			t.Errorf("Expected mode %q to be valid, but VerifyValidMode returned error: %v", modeStr, err)
			continue
		}

		// Verify template exists and is non-empty
		template, exists := PromptTemplates[PromptMode(mode)]
		if !exists {
			t.Errorf("Expected PromptTemplates to contain mode %q", modeStr)
			continue
		}

		if template == "" {
			t.Errorf("Expected template for mode %q to be non-empty", modeStr)
		}
	}
}
