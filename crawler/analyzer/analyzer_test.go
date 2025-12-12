package analyzer

import (
	"encoding/json"
	"testing"
)

func TestParseJSONResponse_PlainJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "plain JSON object",
			input:    `{"is_joke": true, "confidence": 85, "reasoning": "This is clearly a joke"}`,
			expected: `{"is_joke": true, "confidence": 85, "reasoning": "This is clearly a joke"}`,
			wantErr:  false,
		},
		{
			name:     "JSON with whitespace",
			input:    `  {"is_joke": false, "confidence": 90, "reasoning": "Serious article"}  `,
			expected: `{"is_joke": false, "confidence": 90, "reasoning": "Serious article"}`,
			wantErr:  false,
		},
		{
			name:     "JSON with newlines",
			input:    "{\n  \"is_joke\": true,\n  \"confidence\": 50,\n  \"reasoning\": \"Hard to tell\"\n}",
			expected: "{\n  \"is_joke\": true,\n  \"confidence\": 50,\n  \"reasoning\": \"Hard to tell\"\n}",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseJSONResponse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseJSONResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("parseJSONResponse() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseJSONResponse_MarkdownCodeBlocks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "JSON wrapped in ```json code block",
			input:    "```json\n{\"is_joke\": true, \"confidence\": 75, \"reasoning\": \"Test\"}\n```",
			expected: `{"is_joke": true, "confidence": 75, "reasoning": "Test"}`,
			wantErr:  false,
		},
		{
			name:     "JSON wrapped in ``` code block",
			input:    "```\n{\"is_joke\": false, \"confidence\": 95, \"reasoning\": \"Serious\"}\n```",
			expected: `{"is_joke": false, "confidence": 95, "reasoning": "Serious"}`,
			wantErr:  false,
		},
		{
			name:     "JSON in code block with extra whitespace",
			input:    "```json\n  {\"is_joke\": true, \"confidence\": 60, \"reasoning\": \"Maybe\"}  \n```",
			expected: `{"is_joke": true, "confidence": 60, "reasoning": "Maybe"}`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseJSONResponse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseJSONResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Compare normalized JSON (parse and re-marshal to handle whitespace differences)
				var gotObj, expectedObj map[string]interface{}
				if err := json.Unmarshal([]byte(result), &gotObj); err != nil {
					t.Errorf("parsed result is not valid JSON: %v", err)
					return
				}
				if err := json.Unmarshal([]byte(tt.expected), &expectedObj); err != nil {
					t.Errorf("expected result is not valid JSON: %v", err)
					return
				}
				gotJSON, _ := json.Marshal(gotObj)
				expectedJSON, _ := json.Marshal(expectedObj)
				if string(gotJSON) != string(expectedJSON) {
					t.Errorf("parseJSONResponse() = %q, want %q", result, tt.expected)
				}
			}
		})
	}
}

func TestParseJSONResponse_WithExtraText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		validate func(t *testing.T, result string)
	}{
		{
			name:    "JSON with text before",
			input:   "Here is the analysis: {\"is_joke\": true, \"confidence\": 80, \"reasoning\": \"Funny\"}",
			wantErr: false,
			validate: func(t *testing.T, result string) {
				var obj AnalysisResult
				if err := json.Unmarshal([]byte(result), &obj); err != nil {
					t.Errorf("result is not valid JSON: %v", err)
					return
				}
				if obj.IsJoke != true {
					t.Errorf("expected IsJoke = true, got %v", obj.IsJoke)
				}
			},
		},
		{
			name:    "JSON with text after",
			input:   "{\"is_joke\": false, \"confidence\": 90, \"reasoning\": \"Real\"} That's my analysis.",
			wantErr: false,
			validate: func(t *testing.T, result string) {
				var obj AnalysisResult
				if err := json.Unmarshal([]byte(result), &obj); err != nil {
					t.Errorf("result is not valid JSON: %v", err)
					return
				}
				if obj.IsJoke != false {
					t.Errorf("expected IsJoke = false, got %v", obj.IsJoke)
				}
			},
		},
		{
			name:    "JSON with text before and after",
			input:   "Analysis result: {\"is_joke\": true, \"confidence\": 55, \"reasoning\": \"Unclear\"} End of analysis.",
			wantErr: false,
			validate: func(t *testing.T, result string) {
				var obj AnalysisResult
				if err := json.Unmarshal([]byte(result), &obj); err != nil {
					t.Errorf("result is not valid JSON: %v", err)
					return
				}
				if obj.IsJoke != true {
					t.Errorf("expected IsJoke = true, got %v", obj.IsJoke)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseJSONResponse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseJSONResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestParseJSONResponse_InvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "no JSON object",
			input:   "This is just text with no JSON",
			wantErr: true,
		},
		{
			name:    "only opening brace",
			input:   "{",
			wantErr: true,
		},
		{
			name:    "only closing brace",
			input:   "}",
			wantErr: true,
		},
		{
			name:    "closing brace before opening",
			input:   "} {",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   \n\t  ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseJSONResponse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseJSONResponse() error = %v, wantErr %v. Result: %q", err, tt.wantErr, result)
			}
			if !tt.wantErr && result == "" {
				t.Error("parseJSONResponse() returned empty result without error")
			}
		})
	}
}

func TestParseJSONResponse_UnmarshalToAnalysisResult(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedResult AnalysisResult
		wantErr        bool
	}{
		{
			name:  "valid JSON with all fields",
			input: `{"is_joke": true, "confidence": 85, "reasoning": "This is clearly a joke"}`,
			expectedResult: AnalysisResult{
				IsJoke:     true,
				Confidence: 85,
				Reasoning:  "This is clearly a joke",
			},
			wantErr: false,
		},
		{
			name:  "valid JSON with false",
			input: `{"is_joke": false, "confidence": 90, "reasoning": "Serious article"}`,
			expectedResult: AnalysisResult{
				IsJoke:     false,
				Confidence: 90,
				Reasoning:  "Serious article",
			},
			wantErr: false,
		},
		{
			name:  "valid JSON with true",
			input: `{"is_joke": true, "confidence": 50, "reasoning": "Hard to determine"}`,
			expectedResult: AnalysisResult{
				IsJoke:     true,
				Confidence: 50,
				Reasoning:  "Hard to determine",
			},
			wantErr: false,
		},
		{
			name:  "JSON in markdown code block",
			input: "```json\n{\"is_joke\": true, \"confidence\": 75, \"reasoning\": \"Funny content\"}\n```",
			expectedResult: AnalysisResult{
				IsJoke:     true,
				Confidence: 75,
				Reasoning:  "Funny content",
			},
			wantErr: false,
		},
		{
			name:  "JSON with extra text",
			input: "Here's the result: {\"is_joke\": false, \"confidence\": 95, \"reasoning\": \"Real news\"} That's it.",
			expectedResult: AnalysisResult{
				IsJoke:     false,
				Confidence: 95,
				Reasoning:  "Real news",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonStr, err := parseJSONResponse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseJSONResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			var result AnalysisResult
			if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
				t.Errorf("failed to unmarshal parsed JSON: %v. JSON: %q", err, jsonStr)
				return
			}

			if result.IsJoke != tt.expectedResult.IsJoke {
				t.Errorf("IsJoke = %v, want %v", result.IsJoke, tt.expectedResult.IsJoke)
			}
			if result.Confidence != tt.expectedResult.Confidence {
				t.Errorf("Confidence = %d, want %d", result.Confidence, tt.expectedResult.Confidence)
			}
			if result.Reasoning != tt.expectedResult.Reasoning {
				t.Errorf("Reasoning = %q, want %q", result.Reasoning, tt.expectedResult.Reasoning)
			}
		})
	}
}

func TestParseJSONResponse_EdgeCases(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		wantErr            bool
		skipJSONValidation bool
	}{
		{
			name:    "nested JSON objects",
			input:   `{"is_joke": true, "confidence": 80, "reasoning": "Test", "metadata": {"source": "test"}}`,
			wantErr: false,
		},
		{
			name:    "JSON with escaped quotes in reasoning",
			input:   `{"is_joke": true, "confidence": 85, "reasoning": "This is a \"joke\" article"}`,
			wantErr: false,
		},
		{
			name:    "JSON with unicode characters",
			input:   `{"is_joke": true, "confidence": 70, "reasoning": "This is a joke 😂"}`,
			wantErr: false,
		},
		{
			name:               "multiple JSON objects (extracts from first { to last }, may be invalid)",
			input:              `{"is_joke": true, "confidence": 80, "reasoning": "First"} {"is_joke": false, "confidence": 90, "reasoning": "Second"}`,
			wantErr:            false,
			skipJSONValidation: true, // This will extract both objects as one, which is invalid JSON
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseJSONResponse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseJSONResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !tt.skipJSONValidation {
				// Verify it's valid JSON
				var obj map[string]interface{}
				if err := json.Unmarshal([]byte(result), &obj); err != nil {
					t.Errorf("parsed result is not valid JSON: %v. Result: %q", err, result)
					return
				}
				// Verify it contains expected fields
				if _, ok := obj["is_joke"]; !ok {
					t.Error("parsed JSON missing 'is_joke' field")
				}
				if _, ok := obj["confidence"]; !ok {
					t.Error("parsed JSON missing 'confidence' field")
				}
				if _, ok := obj["reasoning"]; !ok {
					t.Error("parsed JSON missing 'reasoning' field")
				}
			}
		})
	}
}

func TestParseJSONResponse_RealWorldExamples(t *testing.T) {
	// Test cases that might come from actual LLM responses
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, result string)
	}{
		{
			name: "LLM response with explanation before JSON",
			input: `Based on my analysis, here is the result:
{
  "is_joke": true,
  "confidence": 85,
  "reasoning": "The article contains satirical elements"
}`,
			wantErr: false,
			check: func(t *testing.T, result string) {
				var obj AnalysisResult
				if err := json.Unmarshal([]byte(result), &obj); err != nil {
					t.Errorf("failed to unmarshal: %v", err)
					return
				}
				if obj.IsJoke != true || obj.Confidence != 85 {
					t.Errorf("unexpected values: IsJoke=%v, Confidence=%d", obj.IsJoke, obj.Confidence)
				}
			},
		},
		{
			name:    "LLM response in code block with newlines",
			input:   "```json\n{\n  \"is_joke\": false,\n  \"confidence\": 92,\n  \"reasoning\": \"This is a serious news article\"\n}\n```",
			wantErr: false,
			check: func(t *testing.T, result string) {
				var obj AnalysisResult
				if err := json.Unmarshal([]byte(result), &obj); err != nil {
					t.Errorf("failed to unmarshal: %v", err)
					return
				}
				if obj.IsJoke != false || obj.Confidence != 92 {
					t.Errorf("unexpected values: IsJoke=%v, Confidence=%d", obj.IsJoke, obj.Confidence)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseJSONResponse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseJSONResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}
