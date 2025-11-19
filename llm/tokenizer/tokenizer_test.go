package tokenizer

import (
	"testing"
)

func TestNewTokenCounter(t *testing.T) {
	tc, err := NewTokenCounter()
	if err != nil {
		t.Fatalf("NewTokenCounter() failed: %v", err)
	}
	if tc == nil {
		t.Fatal("NewTokenCounter() returned nil TokenCounter")
	}
	if tc.encoding == nil {
		t.Fatal("NewTokenCounter() returned TokenCounter with nil encoding")
	}
}

func TestCountTokens(t *testing.T) {
	tc, err := NewTokenCounter()
	if err != nil {
		t.Fatalf("NewTokenCounter() failed: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "simple word",
			input:    "hello",
			expected: 1,
		},
		{
			name:     "simple sentence",
			input:    "Hello, world!",
			expected: 4,
		},
		{
			name:     "longer text",
			input:    "The quick brown fox jumps over the lazy dog.",
			expected: 10,
		},
		{
			name:     "technical text",
			input:    "func main() { fmt.Println(\"Hello\") }",
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := tc.CountTokens(tt.input)
			if count != tt.expected {
				t.Errorf("CountTokens(%q) = %d, expected %d", tt.input, count, tt.expected)
			}
		})
	}
}

func TestCountTokensNilEncoding(t *testing.T) {
	tc := &TokenCounter{encoding: nil}
	count := tc.CountTokens("hello")
	if count != 0 {
		t.Errorf("CountTokens with nil encoding = %d, expected 0", count)
	}
}

func TestCountTokensMultiple(t *testing.T) {
	tc, err := NewTokenCounter()
	if err != nil {
		t.Fatalf("NewTokenCounter() failed: %v", err)
	}

	tests := []struct {
		name     string
		inputs   []string
		expected int
	}{
		{
			name:     "empty slice",
			inputs:   []string{},
			expected: 0,
		},
		{
			name:     "single text",
			inputs:   []string{"hello"},
			expected: 1,
		},
		{
			name:     "multiple texts",
			inputs:   []string{"Hello", "world", "!"},
			expected: 3,
		},
		{
			name:     "mixed lengths",
			inputs:   []string{"The quick brown fox", "jumps over", "the lazy dog."},
			expected: 11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := tc.CountTokensMultiple(tt.inputs)
			if count != tt.expected {
				t.Errorf("CountTokensMultiple(%v) = %d, expected %d", tt.inputs, count, tt.expected)
			}
		})
	}
}

func TestCalculateBYOWCost(t *testing.T) {
	tc, err := NewTokenCounter()
	if err != nil {
		t.Fatalf("NewTokenCounter() failed: %v", err)
	}

	tests := []struct {
		name         string
		inputTokens  int64
		outputTokens int64
		expected     float64
	}{
		{
			name:         "zero tokens",
			inputTokens:  0,
			outputTokens: 0,
			expected:     0.0,
		},
		{
			name:         "1M input tokens only",
			inputTokens:  1_000_000,
			outputTokens: 0,
			expected:     0.03,
		},
		{
			name:         "1M output tokens only",
			inputTokens:  0,
			outputTokens: 1_000_000,
			expected:     0.12,
		},
		{
			name:         "1M input + 1M output",
			inputTokens:  1_000_000,
			outputTokens: 1_000_000,
			expected:     0.15,
		},
		{
			name:         "500K input + 250K output",
			inputTokens:  500_000,
			outputTokens: 250_000,
			expected:     0.045, // (500K / 1M * 0.03) + (250K / 1M * 0.12) = 0.015 + 0.03 = 0.045
		},
		{
			name:         "small amounts",
			inputTokens:  1000,
			outputTokens: 500,
			expected:     0.00009, // (1000 / 1M * 0.03) + (500 / 1M * 0.12) = 0.00003 + 0.00006 = 0.00009
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := tc.CalculateBYOWCost(tt.inputTokens, tt.outputTokens)
			// Allow small floating point differences
			epsilon := 0.0000001
			if cost < tt.expected-epsilon || cost > tt.expected+epsilon {
				t.Errorf("CalculateBYOWCost(%d, %d) = %f, expected %f", tt.inputTokens, tt.outputTokens, cost, tt.expected)
			}
		})
	}
}

func TestCountTokensGlobal(t *testing.T) {
	// Reset global counter to ensure clean test
	globalTokenCounter = nil

	tests := []struct {
		name     string
		input    string
		minCount int // Use minimum count since exact counts may vary
	}{
		{
			name:     "empty string",
			input:    "",
			minCount: 0,
		},
		{
			name:     "simple word",
			input:    "hello",
			minCount: 1,
		},
		{
			name:     "sentence",
			input:    "Hello, world!",
			minCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := CountTokensGlobal(tt.input)
			if count < tt.minCount {
				t.Errorf("CountTokensGlobal(%q) = %d, expected at least %d", tt.input, count, tt.minCount)
			}
		})
	}

	// Verify singleton behavior - calling again should use same instance
	count1 := CountTokensGlobal("test")
	count2 := CountTokensGlobal("test")
	if count1 != count2 {
		t.Errorf("CountTokensGlobal singleton behavior failed: got different counts %d vs %d", count1, count2)
	}
}

func BenchmarkCountTokens(b *testing.B) {
	tc, err := NewTokenCounter()
	if err != nil {
		b.Fatalf("NewTokenCounter() failed: %v", err)
	}

	text := "The quick brown fox jumps over the lazy dog. " +
		"This is a benchmark test for token counting performance. " +
		"We want to ensure the token counter is efficient for production use."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tc.CountTokens(text)
	}
}

func BenchmarkCountTokensGlobal(b *testing.B) {
	text := "The quick brown fox jumps over the lazy dog."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CountTokensGlobal(text)
	}
}
