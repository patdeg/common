package llmutils

import (
	"reflect"
	"strings"
	"testing"
)

// TestProcess tests the main processing function with metadata extraction
func TestProcess(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedCleaned string
		expectedParams  map[string]string
		expectedFlow    string
		expectedNode    string
		expectedTags    []string
	}{
		{
			name: "full line comments",
			input: `/// This is a comment
You are a helpful assistant
/// Another comment
Be concise`,
			expectedCleaned: `You are a helpful assistant
Be concise`,
			expectedParams: map[string]string{},
			expectedFlow:   "",
			expectedNode:   "",
			expectedTags:   []string{},
		},
		{
			name: "inline comments",
			input: `You are a helpful assistant /// inline comment
Be concise /// another inline comment`,
			expectedCleaned: `You are a helpful assistant
Be concise`,
			expectedParams: map[string]string{},
			expectedFlow:   "",
			expectedNode:   "",
			expectedTags:   []string{},
		},
		{
			name: "param extraction single",
			input: `/// param: model=gpt-4
You are a helpful assistant`,
			expectedCleaned: `You are a helpful assistant`,
			expectedParams:  map[string]string{"model": "gpt-4"},
			expectedFlow:    "",
			expectedNode:    "",
			expectedTags:    []string{},
		},
		{
			name: "param extraction multiple on same line",
			input: `/// param: model=gpt-4, temperature=0.7, max_tokens=1000
You are a helpful assistant`,
			expectedCleaned: `You are a helpful assistant`,
			expectedParams: map[string]string{
				"model":       "gpt-4",
				"temperature": "0.7",
				"max_tokens":  "1000",
			},
			expectedFlow: "",
			expectedNode: "",
			expectedTags: []string{},
		},
		{
			name: "param extraction multiple lines",
			input: `/// param: model=gpt-4
/// param: temperature=0.7
You are a helpful assistant
/// param: max_tokens=1000`,
			expectedCleaned: `You are a helpful assistant`,
			expectedParams: map[string]string{
				"model":       "gpt-4",
				"temperature": "0.7",
				"max_tokens":  "1000",
			},
			expectedFlow: "",
			expectedNode: "",
			expectedTags: []string{},
		},
		{
			name: "flow and node extraction",
			input: `/// flow: checkout-process
/// node: payment-validation
You are a helpful assistant`,
			expectedCleaned: `You are a helpful assistant`,
			expectedParams:  map[string]string{},
			expectedFlow:    "checkout-process",
			expectedNode:    "payment-validation",
			expectedTags:    []string{"flow:checkout-process", "node:payment-validation"},
		},
		{
			name: "tags extraction",
			input: `/// tags: important, production
You are a helpful assistant`,
			expectedCleaned: `You are a helpful assistant`,
			expectedParams:  map[string]string{},
			expectedFlow:    "",
			expectedNode:    "",
			expectedTags:    []string{"important", "production"},
		},
		{
			name: "mixed comments and params",
			input: `/// This is a documentation comment
/// param: model=gpt-4
/// flow: checkout
You are a helpful assistant /// inline comment
/// Another doc comment
/// param: temperature=0.7
Be concise`,
			expectedCleaned: `You are a helpful assistant
Be concise`,
			expectedParams: map[string]string{
				"model":       "gpt-4",
				"temperature": "0.7",
			},
			expectedFlow: "checkout",
			expectedNode: "",
			expectedTags: []string{"flow:checkout"},
		},
		{
			name: "whitespace handling",
			input: `/// param: model = gpt-4 , temperature = 0.7
You are a helpful assistant   /// inline with trailing spaces`,
			expectedCleaned: `You are a helpful assistant`,
			expectedParams: map[string]string{
				"model":       "gpt-4",
				"temperature": "0.7",
			},
			expectedFlow: "",
			expectedNode: "",
			expectedTags: []string{},
		},
		{
			name: "empty lines preserved",
			input: `/// Comment

You are a helpful assistant

Be concise`,
			expectedCleaned: `
You are a helpful assistant

Be concise`,
			expectedParams: map[string]string{},
			expectedFlow:   "",
			expectedNode:   "",
			expectedTags:   []string{},
		},
		{
			name: "no comments",
			input: `You are a helpful assistant
Be concise`,
			expectedCleaned: `You are a helpful assistant
Be concise`,
			expectedParams: map[string]string{},
			expectedFlow:   "",
			expectedNode:   "",
			expectedTags:   []string{},
		},
		{
			name:            "empty string",
			input:           "",
			expectedCleaned: "",
			expectedParams:  map[string]string{},
			expectedFlow:    "",
			expectedNode:    "",
			expectedTags:    []string{},
		},
		{
			name: "only comments",
			input: `/// Comment 1
/// param: model=gpt-4
/// Comment 2`,
			expectedCleaned: "",
			expectedParams:  map[string]string{"model": "gpt-4"},
			expectedFlow:    "",
			expectedNode:    "",
			expectedTags:    []string{},
		},
		{
			name: "param with special characters",
			input: `/// param: prompt_id=cv_system_v2.1
/// param: version=2024-10-11
You are a helpful assistant`,
			expectedCleaned: `You are a helpful assistant`,
			expectedParams: map[string]string{
				"prompt_id": "cv_system_v2.1",
				"version":   "2024-10-11",
			},
			expectedFlow: "",
			expectedNode: "",
			expectedTags: []string{},
		},
		{
			name: "URLs are not treated as comments",
			input: `Visit http://example.com for more info
Also check https://google.com`,
			expectedCleaned: `Visit http://example.com for more info
Also check https://google.com`,
			expectedParams: map[string]string{},
			expectedFlow:   "",
			expectedNode:   "",
			expectedTags:   []string{},
		},
		{
			name: "URL with comment after",
			input: `Visit http://example.com /// great site
Also https://google.com /// search engine`,
			expectedCleaned: `Visit http://example.com
Also https://google.com`,
			expectedParams: map[string]string{},
			expectedFlow:   "",
			expectedNode:   "",
			expectedTags:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Process(tt.input)

			if result.CleanedPrompt != tt.expectedCleaned {
				t.Errorf("CleanedPrompt mismatch\nGot:\n%q\n\nWant:\n%q", result.CleanedPrompt, tt.expectedCleaned)
			}

			if !reflect.DeepEqual(result.Params, tt.expectedParams) {
				t.Errorf("Params mismatch\nGot:  %v\nWant: %v", result.Params, tt.expectedParams)
			}

			if result.Flow != tt.expectedFlow {
				t.Errorf("Flow mismatch\nGot:  %q\nWant: %q", result.Flow, tt.expectedFlow)
			}

			if result.Node != tt.expectedNode {
				t.Errorf("Node mismatch\nGot:  %q\nWant: %q", result.Node, tt.expectedNode)
			}

			if !reflect.DeepEqual(result.Tags, tt.expectedTags) {
				t.Errorf("Tags mismatch\nGot:  %v\nWant: %v", result.Tags, tt.expectedTags)
			}
		})
	}
}

// TestStripComments tests the comment stripping function
func TestStripComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No comments",
			input:    "Hello world",
			expected: "Hello world",
		},
		{
			name:     "Inline comment",
			input:    "Hello world /// this is a comment",
			expected: "Hello world",
		},
		{
			name:     "Full-line comment",
			input:    "/// This is a comment",
			expected: "",
		},
		{
			name:     "Full-line comment with leading whitespace",
			input:    "   /// This is a comment",
			expected: "",
		},
		{
			name:     "Multiple lines with inline comments",
			input:    "Line 1 /// comment 1\nLine 2 /// comment 2\nLine 3",
			expected: "Line 1\nLine 2\nLine 3",
		},
		{
			name:     "Mixed full-line and inline comments",
			input:    "/// Header comment\nActual text /// inline\n/// Footer comment",
			expected: "Actual text",
		},
		{
			name: "Multi-line with blank line removal",
			input: `First line /// comment
/// Full line comment

Second line`,
			expected: "First line\nSecond line",
		},
		{
			name:     "URL with http protocol",
			input:    "Visit http://example.com for info",
			expected: "Visit http://example.com for info",
		},
		{
			name:     "URL with https protocol",
			input:    "Visit https://example.com for info",
			expected: "Visit https://example.com for info",
		},
		{
			name:     "URL with comment after",
			input:    "Visit http://example.com /// great site",
			expected: "Visit http://example.com",
		},
		{
			name:     "Multiple URLs",
			input:    "http://example.com and https://google.com are sites",
			expected: "http://example.com and https://google.com are sites",
		},
		{
			name:     "URL at end of line",
			input:    "Check out http://example.com",
			expected: "Check out http://example.com",
		},
		{
			name:     "Comment looks like URL but isn't",
			input:    "Three slashes: /// not a url",
			expected: "Three slashes:",
		},
		{
			name:     "Empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "Only comments",
			input:    "/// Comment 1\n/// Comment 2\n/// Comment 3",
			expected: "",
		},
		{
			name:     "Comment with special characters",
			input:    "Text /// comment with !@#$%^&*() special chars",
			expected: "Text",
		},
		{
			name: "Real-world prompt example",
			input: `You are a helpful assistant. /// Be polite
When answering questions:
1. Be clear /// avoid jargon
2. Be concise /// no rambling
3. Be accurate

/// Debug: This is a test prompt
Always cite sources when possible.`,
			expected: `You are a helpful assistant.
When answering questions:
1. Be clear
2. Be concise
3. Be accurate
Always cite sources when possible.`,
		},
		{
			name:     "Tabs and spaces before comment",
			input:    "Text\t/// comment with tab",
			expected: "Text",
		},
		{
			name:     "Multiple slashes not at start",
			input:    "C://Windows/System32 is a path",
			expected: "C://Windows/System32 is a path",
		},
		{
			name:     "Triple slash in middle of text (not comment)",
			input:    "The pattern /// matches comments",
			expected: "The pattern",
		},
		{
			name:     "Only whitespace",
			input:    "   \t  \n  ",
			expected: "",
		},
		{
			name:     "Unicode characters before comment",
			input:    "Hello 世界 /// comment",
			expected: "Hello 世界",
		},
		{
			name:     "Emoji before comment",
			input:    "Happy 😀 /// smile",
			expected: "Happy 😀",
		},
		{
			name:     "Very long line with comment",
			input:    strings.Repeat("a", 10000) + " /// comment",
			expected: strings.Repeat("a", 10000),
		},
		{
			name:     "Comment with only spaces after marker",
			input:    "Text ///    ",
			expected: "Text",
		},
		{
			name:     "Multiple comment markers",
			input:    "Text /// comment 1 /// comment 2",
			expected: "Text", // Only first /// counts
		},
		{
			name:     "Nested slashes",
			input:    "Text /// /// nested",
			expected: "Text",
		},
		{
			name:     "Windows file path",
			input:    "C://Windows/System32",
			expected: "C://Windows/System32", // // not ///
		},
		{
			name:     "Unix-style comment (not triple slash)",
			input:    "Text // not a comment",
			expected: "Text // not a comment",
		},
		{
			name:     "URL with port",
			input:    "http://example.com:8080/path",
			expected: "http://example.com:8080/path",
		},
		{
			name:     "URL with query params",
			input:    "http://example.com?foo=bar&baz=qux",
			expected: "http://example.com?foo=bar&baz=qux",
		},
		{
			name:     "FTP protocol (not http/https)",
			input:    "ftp://example.com /// comment",
			expected: "ftp://example.com", // ftp:// is still URL-like
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripComments(tt.input)

			if result != tt.expected {
				t.Errorf("StripComments() failed\nInput:\n%s\n\nExpected:\n%s\n\nGot:\n%s",
					tt.input, tt.expected, result)
			}
		})
	}
}

// TestExtractParams tests the param extraction function
func TestExtractParams(t *testing.T) {
	input := `/// This is a comment
/// param: model=gpt-4
You are a helpful assistant
/// param: temperature=0.7`

	expected := map[string]string{
		"model":       "gpt-4",
		"temperature": "0.7",
	}

	result := ExtractParams(input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("ExtractParams mismatch\nGot:  %v\nWant: %v", result, expected)
	}
}

// TestStripCommentFromLine tests comment stripping on individual lines
func TestStripCommentFromLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No comment",
			input:    "Hello world",
			expected: "Hello world",
		},
		{
			name:     "Inline comment",
			input:    "Text /// comment",
			expected: "Text",
		},
		{
			name:     "Full-line comment",
			input:    "/// Full comment",
			expected: "",
		},
		{
			name:     "Comment with leading spaces",
			input:    "   /// Comment",
			expected: "",
		},
		{
			name:     "Comment with leading tabs",
			input:    "\t\t/// Comment",
			expected: "",
		},
		{
			name:     "URL not treated as comment",
			input:    "http://example.com",
			expected: "http://example.com",
		},
		{
			name:     "HTTPS URL not treated as comment",
			input:    "https://example.com",
			expected: "https://example.com",
		},
		{
			name:     "URL with path",
			input:    "http://example.com/path/to/resource",
			expected: "http://example.com/path/to/resource",
		},
		{
			name:     "Text before and after URL",
			input:    "Visit http://example.com for more",
			expected: "Visit http://example.com for more",
		},
		{
			name:     "URL followed by comment",
			input:    "http://example.com /// good site",
			expected: "http://example.com",
		},
		{
			name:     "Multiple URLs in line",
			input:    "http://a.com and https://b.com",
			expected: "http://a.com and https://b.com",
		},
		{
			name:     "Trailing whitespace before comment",
			input:    "Text   /// comment",
			expected: "Text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripCommentFromLine(tt.input)

			if result != tt.expected {
				t.Errorf("stripCommentFromLine() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestFindCommentPosition tests the comment position detection
func TestFindCommentPosition(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "No slashes",
			input:    "Hello world",
			expected: -1,
		},
		{
			name:     "Comment at start",
			input:    "/// comment",
			expected: 0,
		},
		{
			name:     "Comment in middle",
			input:    "Text /// comment",
			expected: 5,
		},
		{
			name:     "Comment at end",
			input:    "Text ///",
			expected: 5,
		},
		{
			name:     "URL with http",
			input:    "http://example.com",
			expected: -1, // No comment, /// is part of URL
		},
		{
			name:     "URL with https",
			input:    "https://example.com",
			expected: -1, // No comment
		},
		{
			name:     "URL then comment",
			input:    "http://example.com /// comment",
			expected: 19, // Position after URL
		},
		{
			name:     "Multiple URLs no comment",
			input:    "http://a.com https://b.com",
			expected: -1,
		},
		{
			name:     "Two slashes (not three)",
			input:    "Text // not comment",
			expected: -1,
		},
		{
			name:     "Four slashes (contains three)",
			input:    "Text //// comment",
			expected: 5, // First /// at position 5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findCommentPosition(tt.input)

			if result != tt.expected {
				t.Errorf("findCommentPosition(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

// TestIsPartOfURL tests URL detection logic
func TestIsPartOfURL(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		slashPos  int
		expectURL bool
	}{
		{
			name:      "http protocol",
			line:      "http://example.com",
			slashPos:  5, // Position of first / in ://
			expectURL: true,
		},
		{
			name:      "https protocol",
			line:      "https://example.com",
			slashPos:  6, // Position of first / in ://
			expectURL: true,
		},
		{
			name:      "Not a URL",
			line:      "text /// comment",
			slashPos:  5,
			expectURL: false,
		},
		{
			name:      "URL in middle of text",
			line:      "Visit http://example.com here",
			slashPos:  11, // Position of /// in URL
			expectURL: true,
		},
		{
			name:      "Comment after URL",
			line:      "http://example.com /// comment",
			slashPos:  19, // Position of /// comment marker
			expectURL: false,
		},
		{
			name:      "Position too early for protocol",
			line:      "/// comment",
			slashPos:  0, // No room for http: before
			expectURL: false,
		},
		{
			name:      "Partial protocol",
			line:      "htt://example.com",
			slashPos:  4,
			expectURL: false, // Not http: or https:
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPartOfURL(tt.line, tt.slashPos)

			if result != tt.expectURL {
				t.Errorf("isPartOfURL(%q, %d) = %v, want %v",
					tt.line, tt.slashPos, result, tt.expectURL)
			}
		})
	}
}

// TestStripCommentsFromMessages tests message array processing
func TestStripCommentsFromMessages(t *testing.T) {
	tests := []struct {
		name     string
		messages []interface{}
		expected []interface{}
	}{
		{
			name: "Single message with inline comment",
			messages: []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Hello world /// comment",
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Hello world",
				},
			},
		},
		{
			name: "Multiple messages with comments",
			messages: []interface{}{
				map[string]interface{}{
					"role":    "system",
					"content": "You are helpful /// be nice",
				},
				map[string]interface{}{
					"role":    "user",
					"content": "Tell me a joke /// make it funny",
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"role":    "system",
					"content": "You are helpful",
				},
				map[string]interface{}{
					"role":    "user",
					"content": "Tell me a joke",
				},
			},
		},
		{
			name: "Message with full-line comment",
			messages: []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "/// Debug note\nActual question here",
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Actual question here",
				},
			},
		},
		{
			name: "Message without comments",
			messages: []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Simple question",
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Simple question",
				},
			},
		},
		{
			name: "Message with URL",
			messages: []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Check http://example.com /// good site",
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Check http://example.com",
				},
			},
		},
		{
			name: "Message with additional fields",
			messages: []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Question /// comment",
					"name":    "TestUser",
					"metadata": map[string]string{
						"source": "test",
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"role":    "user",
					"content": "Question",
					"name":    "TestUser",
					"metadata": map[string]string{
						"source": "test",
					},
				},
			},
		},
		{
			name:     "Empty messages array",
			messages: []interface{}{},
			expected: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripCommentsFromMessages(tt.messages)

			if len(result) != len(tt.expected) {
				t.Fatalf("Result length = %d, want %d", len(result), len(tt.expected))
			}

			for i := range result {
				resultMsg, ok1 := result[i].(map[string]interface{})
				expectedMsg, ok2 := tt.expected[i].(map[string]interface{})

				if !ok1 || !ok2 {
					t.Fatalf("Message %d is not a map", i)
				}

				// Check role
				if resultMsg["role"] != expectedMsg["role"] {
					t.Errorf("Message %d role = %v, want %v",
						i, resultMsg["role"], expectedMsg["role"])
				}

				// Check content
				if resultMsg["content"] != expectedMsg["content"] {
					t.Errorf("Message %d content = %q, want %q",
						i, resultMsg["content"], expectedMsg["content"])
				}

				// Check that other fields are preserved (count only)
				for key := range expectedMsg {
					if key != "role" && key != "content" {
						if _, exists := resultMsg[key]; !exists {
							t.Errorf("Message %d missing field %s", i, key)
						}
					}
				}
			}
		})
	}
}

// BenchmarkProcess benchmarks the full processing with metadata extraction
func BenchmarkProcess(b *testing.B) {
	input := `/// This is a comment
/// param: model=gpt-4, temperature=0.7, max_tokens=1000
/// flow: checkout
/// node: payment
You are a helpful assistant /// inline comment
/// Another comment
Be concise and helpful
/// param: version=1.0`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Process(input)
	}
}

// BenchmarkStripCommentsOnly benchmarks comment stripping without metadata
func BenchmarkStripCommentsOnly(b *testing.B) {
	input := `/// This is a comment
You are a helpful assistant /// inline comment
/// Another comment
Be concise and helpful`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = StripComments(input)
	}
}

// BenchmarkStripComments benchmarks the main comment stripping with blank line removal
func BenchmarkStripComments(b *testing.B) {
	input := `You are a helpful assistant. /// Be polite
When answering questions:
1. Be clear /// avoid jargon
2. Be concise /// no rambling
3. Be accurate

/// Debug: This is a test prompt
Always cite sources when possible.
Visit http://example.com for more info /// great site`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = StripComments(input)
	}
}

// BenchmarkStripCommentsFromMessages benchmarks message array processing
func BenchmarkStripCommentsFromMessages(b *testing.B) {
	messages := []interface{}{
		map[string]interface{}{
			"role":    "system",
			"content": "You are helpful /// be nice",
		},
		map[string]interface{}{
			"role":    "user",
			"content": "Tell me a joke /// make it funny\n/// Debug note\nActual question",
		},
		map[string]interface{}{
			"role":    "assistant",
			"content": "Here's a joke /// no comment",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = StripCommentsFromMessages(messages)
	}
}
