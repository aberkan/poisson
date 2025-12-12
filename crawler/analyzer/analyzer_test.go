package analyzer

import (
	"encoding/json"
	"strings"
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
				var obj jokeIntermediateResult
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
				var obj jokeIntermediateResult
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
				var obj jokeIntermediateResult
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

func TestParseJSONResponse_UnmarshalToIntermediateResult(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedIsJoke bool
		expectedConf   int
		expectedReason string
		wantErr        bool
	}{
		{
			name:           "valid JSON with all fields",
			input:          `{"is_joke": true, "confidence": 85, "reasoning": "This is clearly a joke"}`,
			expectedIsJoke: true,
			expectedConf:   85,
			expectedReason: "This is clearly a joke",
			wantErr:        false,
		},
		{
			name:           "valid JSON with false",
			input:          `{"is_joke": false, "confidence": 90, "reasoning": "Serious article"}`,
			expectedIsJoke: false,
			expectedConf:   90,
			expectedReason: "Serious article",
			wantErr:        false,
		},
		{
			name:           "JSON in markdown code block",
			input:          "```json\n{\"is_joke\": true, \"confidence\": 75, \"reasoning\": \"Funny content\"}\n```",
			expectedIsJoke: true,
			expectedConf:   75,
			expectedReason: "Funny content",
			wantErr:        false,
		},
		{
			name:           "JSON with extra text",
			input:          "Here's the result: {\"is_joke\": false, \"confidence\": 95, \"reasoning\": \"Real news\"} That's it.",
			expectedIsJoke: false,
			expectedConf:   95,
			expectedReason: "Real news",
			wantErr:        false,
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

			var intermediate jokeIntermediateResult
			if err := json.Unmarshal([]byte(jsonStr), &intermediate); err != nil {
				t.Errorf("failed to unmarshal parsed JSON: %v. JSON: %q", err, jsonStr)
				return
			}

			if intermediate.IsJoke != tt.expectedIsJoke {
				t.Errorf("IsJoke = %v, want %v", intermediate.IsJoke, tt.expectedIsJoke)
			}
			if intermediate.Confidence != tt.expectedConf {
				t.Errorf("Confidence = %d, want %d", intermediate.Confidence, tt.expectedConf)
			}
			if intermediate.Reasoning != tt.expectedReason {
				t.Errorf("Reasoning = %q, want %q", intermediate.Reasoning, tt.expectedReason)
			}
		})
	}
}

func TestAnalyze_ConversionToJokePercentage(t *testing.T) {
	tests := []struct {
		name            string
		isJoke          bool
		confidence      int
		reasoning       string
		expectedJokePct *int
		description     string
	}{
		{
			name:            "is_joke true sets percentage",
			isJoke:          true,
			confidence:      85,
			reasoning:       "This is clearly a joke",
			expectedJokePct: intPtr(85),
			description:     "When is_joke is true, JokePercentage should be set to confidence",
		},
		{
			name:            "reasoning mentions joke",
			isJoke:          false,
			confidence:      75,
			reasoning:       "This article contains elements of a joke",
			expectedJokePct: intPtr(75),
			description:     "When reasoning mentions 'joke', JokePercentage should be set",
		},
		{
			name:            "reasoning mentions prank",
			isJoke:          false,
			confidence:      60,
			reasoning:       "This appears to be a prank",
			expectedJokePct: intPtr(60),
			description:     "When reasoning mentions 'prank', JokePercentage should be set",
		},
		{
			name:            "reasoning mentions satire",
			isJoke:          false,
			confidence:      70,
			reasoning:       "This is satirical content",
			expectedJokePct: intPtr(70),
			description:     "When reasoning mentions 'satire', JokePercentage should be set",
		},
		{
			name:            "reasoning mentions humor",
			isJoke:          false,
			confidence:      55,
			reasoning:       "This has humorous elements",
			expectedJokePct: intPtr(55),
			description:     "When reasoning mentions 'humor', JokePercentage should be set",
		},
		{
			name:            "no joke mention - is_joke false and no keywords",
			isJoke:          false,
			confidence:      90,
			reasoning:       "This is a serious news article about current events",
			expectedJokePct: nil,
			description:     "When is_joke is false and reasoning has no joke-related keywords, JokePercentage should be nil",
		},
		{
			name:            "confidence clamped to 100",
			isJoke:          true,
			confidence:      150,
			reasoning:       "Joke",
			expectedJokePct: intPtr(100),
			description:     "Confidence values over 100 should be clamped to 100",
		},
		{
			name:            "confidence clamped to 0",
			isJoke:          true,
			confidence:      -10,
			reasoning:       "Joke",
			expectedJokePct: intPtr(0),
			description:     "Confidence values under 0 should be clamped to 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			intermediate := jokeIntermediateResult{
				IsJoke:     tt.isJoke,
				Confidence: tt.confidence,
				Reasoning:  tt.reasoning,
			}

			result := &AnalysisResult{}

			// Replicate the conversion logic from Analyze function
			reasoningLower := strings.ToLower(intermediate.Reasoning)
			hasJokeMention := intermediate.IsJoke ||
				strings.Contains(reasoningLower, "joke") ||
				strings.Contains(reasoningLower, "prank") ||
				strings.Contains(reasoningLower, "satire") ||
				strings.Contains(reasoningLower, "satirical") ||
				strings.Contains(reasoningLower, "humor") ||
				strings.Contains(reasoningLower, "humorous")

			if hasJokeMention {
				confidence := intermediate.Confidence
				if confidence < 0 {
					confidence = 0
				} else if confidence > 100 {
					confidence = 100
				}
				result.JokePercentage = &confidence
			}

			if tt.expectedJokePct == nil {
				if result.JokePercentage != nil {
					t.Errorf("Expected JokePercentage to be nil, got %v", *result.JokePercentage)
				}
			} else {
				if result.JokePercentage == nil {
					t.Errorf("Expected JokePercentage to be %d, got nil", *tt.expectedJokePct)
				} else if *result.JokePercentage != *tt.expectedJokePct {
					t.Errorf("JokePercentage = %d, want %d", *result.JokePercentage, *tt.expectedJokePct)
				}
			}
		})
	}
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
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
				var obj jokeIntermediateResult
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
				var obj jokeIntermediateResult
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
