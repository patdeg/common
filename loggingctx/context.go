// Package loggingctx provides request ID middleware and context utilities for distributed tracing.
package loggingctx

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"
)

// contextKey is a private type for context keys to avoid collisions.
type contextKey string

const (
	// requestIDKey is the context key for storing request IDs.
	// This matches the ContextKeyRequestID pattern used across services.
	requestIDKey contextKey = "request_id"
)

// GetRequestID extracts the request ID from the given context.
// Returns an empty string if no request ID is found.
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// WithRequestID returns a new context with the given request ID.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// generateULID generates a ULID-like identifier (26 characters, timestamp-prefixed).
// ULID format: 10 bytes timestamp (48 bits milliseconds since epoch) + 16 bytes random
// Encoded in Crockford's base32 (URL-safe, case-insensitive).
//
// This is a simplified ULID implementation that doesn't require external dependencies.
// The format is: TTTTTTTTTTRRRRRRRRRRRRRRRR (10 timestamp chars + 16 random chars)
//
// For full ULID spec compliance, use github.com/oklog/ulid
func generateULID() string {
	// Get current timestamp in milliseconds
	timestamp := uint64(time.Now().UnixMilli())

	// Crockford's Base32 encoding (excludes I, L, O, U to avoid confusion)
	const base32 = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

	// Encode timestamp (48 bits / ~5 bits per char = 10 chars)
	var result [26]byte
	t := timestamp
	for i := 9; i >= 0; i-- {
		result[i] = base32[t&0x1F]
		t >>= 5
	}

	// Generate 16 bytes of random data for the random component
	randomBytes := make([]byte, 10)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return fmt.Sprintf("req_%s_%d",
			time.Now().Format("20060102150405"),
			time.Now().Nanosecond()%1000000)
	}

	// Encode random bytes into base32 (80 bits / ~5 bits per char = 16 chars)
	for i := 0; i < 16; i++ {
		byteIdx := (i * 5) / 8
		bitOffset := (i * 5) % 8

		var value uint16
		if byteIdx < len(randomBytes) {
			value = uint16(randomBytes[byteIdx])
			if byteIdx+1 < len(randomBytes) {
				value = (value << 8) | uint16(randomBytes[byteIdx+1])
			}
		}

		value >>= (11 - bitOffset)
		result[10+i] = base32[value&0x1F]
	}

	return string(result[:])
}
