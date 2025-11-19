package tokenizer

import (
	"fmt"

	"github.com/pkoukk/tiktoken-go"
)

// TokenCounter provides token counting functionality using tiktoken with cl100k_base encoding.
// This encoding is used for:
// - GPT-4, GPT-3.5-turbo models
// - text-embedding-ada-002
// - Webhook/BYOW cost calculation (standardized token counting)
// - Groq models (compatible encoding)
type TokenCounter struct {
	encoding *tiktoken.Tiktoken
}

// NewTokenCounter creates a new token counter using cl100k_base encoding.
// This is the standard encoding for modern OpenAI models and our webhook pricing.
// It's also compatible with Groq models for consistent token counting across providers.
func NewTokenCounter() (*TokenCounter, error) {
	encoding, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return nil, fmt.Errorf("failed to get cl100k_base encoding: %w", err)
	}

	return &TokenCounter{
		encoding: encoding,
	}, nil
}

// CountTokens counts the number of tokens in the given text using cl100k_base encoding.
// This is used for:
// - BYOW (Bring Your Own Workflow) pricing: $0.03 per 1M input tokens, $0.12 per 1M output tokens
// - Accounting/billing purposes
// - Pre-request estimation
func (tc *TokenCounter) CountTokens(text string) int {
	if tc.encoding == nil {
		return 0
	}

	tokens := tc.encoding.Encode(text, nil, nil)
	return len(tokens)
}

// CountTokensMultiple counts tokens for multiple text strings and returns the total.
// Useful for counting tokens across multiple messages in a conversation.
func (tc *TokenCounter) CountTokensMultiple(texts []string) int {
	total := 0
	for _, text := range texts {
		total += tc.CountTokens(text)
	}
	return total
}

// CalculateBYOWCost calculates the cost for Bring Your Own Workflow based on token counts.
// Pricing:
// - Input tokens: $0.03 per 1M tokens
// - Output tokens: $0.12 per 1M tokens
func (tc *TokenCounter) CalculateBYOWCost(inputTokens, outputTokens int64) float64 {
	const (
		inputCostPer1M  = 0.03 // $0.03 per 1M input tokens
		outputCostPer1M = 0.12 // $0.12 per 1M output tokens
	)

	inputCost := float64(inputTokens) / 1_000_000.0 * inputCostPer1M
	outputCost := float64(outputTokens) / 1_000_000.0 * outputCostPer1M

	return inputCost + outputCost
}

// Note: tiktoken-go manages memory automatically via Go's garbage collector.
// No explicit cleanup is required.

// GlobalTokenCounter is a singleton instance for convenient token counting.
// It's initialized lazily on first use.
var globalTokenCounter *TokenCounter

// CountTokensGlobal provides a convenient global function for token counting.
// It uses a singleton TokenCounter instance.
func CountTokensGlobal(text string) int {
	if globalTokenCounter == nil {
		var err error
		globalTokenCounter, err = NewTokenCounter()
		if err != nil {
			return 0
		}
	}
	return globalTokenCounter.CountTokens(text)
}
