package loggingctx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestGetRequestID verifies that GetRequestID extracts request IDs from context.
func TestGetRequestID(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "nil context",
			ctx:      nil,
			expected: "",
		},
		{
			name:     "empty context",
			ctx:      context.Background(),
			expected: "",
		},
		{
			name:     "context with request ID",
			ctx:      WithRequestID(context.Background(), "test-request-id-123"),
			expected: "test-request-id-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRequestID(tt.ctx)
			if result != tt.expected {
				t.Errorf("GetRequestID() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestWithRequestID verifies that WithRequestID stores request IDs in context.
func TestWithRequestID(t *testing.T) {
	ctx := context.Background()
	requestID := "test-request-id-456"

	newCtx := WithRequestID(ctx, requestID)

	result := GetRequestID(newCtx)
	if result != requestID {
		t.Errorf("GetRequestID() = %q, want %q", result, requestID)
	}
}

// TestGenerateULID verifies that generateULID produces valid IDs.
func TestGenerateULID(t *testing.T) {
	// Generate multiple ULIDs to ensure they're unique
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateULID()

		// Check length (should be 26 characters)
		if len(id) != 26 {
			t.Errorf("generateULID() length = %d, want 26", len(id))
		}

		// Check uniqueness
		if ids[id] {
			t.Errorf("generateULID() produced duplicate ID: %s", id)
		}
		ids[id] = true

		// Check that it contains only valid base32 characters
		for _, c := range id {
			if !isValidBase32Char(c) {
				t.Errorf("generateULID() contains invalid character: %c", c)
			}
		}
	}
}

// isValidBase32Char checks if a character is valid in Crockford's Base32.
func isValidBase32Char(c rune) bool {
	return (c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z')
}

// TestRequestIDMiddleware verifies that the middleware generates and propagates request IDs.
func TestRequestIDMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		existingHeader string
		checkGenerated bool
	}{
		{
			name:           "no existing request ID",
			existingHeader: "",
			checkGenerated: true,
		},
		{
			name:           "existing request ID preserved",
			existingHeader: "existing-request-id-789",
			checkGenerated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler that captures context
			var capturedCtx context.Context
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedCtx = r.Context()
				w.WriteHeader(http.StatusOK)
			})

			// Wrap with middleware
			middleware := RequestIDMiddleware(handler)

			// Create test request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.existingHeader != "" {
				req.Header.Set("X-Request-ID", tt.existingHeader)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			middleware.ServeHTTP(rr, req)

			// Verify response header
			responseID := rr.Header().Get("X-Request-ID")
			if responseID == "" {
				t.Error("Response should have X-Request-ID header")
			}

			// Verify context
			contextID := GetRequestID(capturedCtx)
			if contextID == "" {
				t.Error("Context should contain request ID")
			}

			// Verify consistency
			if responseID != contextID {
				t.Errorf("Response header (%s) != context ID (%s)", responseID, contextID)
			}

			// Verify preservation or generation
			if tt.existingHeader != "" {
				if responseID != tt.existingHeader {
					t.Errorf("Existing request ID not preserved: got %s, want %s", responseID, tt.existingHeader)
				}
			} else if tt.checkGenerated {
				if len(responseID) != 26 {
					t.Errorf("Generated request ID should be 26 characters, got %d", len(responseID))
				}
			}
		})
	}
}

// TestLogRequest verifies that LogRequest doesn't panic.
func TestLogRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := WithRequestID(req.Context(), "test-log-request-id")
	req = req.WithContext(ctx)

	// This should not panic
	LogRequest(req, 200, 45*time.Millisecond, 512)
}

// TestLogCall verifies that LogCall doesn't panic with various inputs.
func TestLogCall(t *testing.T) {
	tests := []struct {
		name string
		kv   []interface{}
	}{
		{
			name: "no extra args",
			kv:   nil,
		},
		{
			name: "with extra args",
			kv:   []interface{}{"user_id", "123", "action", "create"},
		},
		{
			name: "odd number of args (should handle gracefully)",
			kv:   []interface{}{"key1", "value1", "key2"},
		},
		{
			name: "non-string key (should skip)",
			kv:   []interface{}{123, "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test?foo=bar", nil)
			req.Header.Set("X-Request-ID", "test-log-call-id")

			// This should not panic
			LogCall(req, "test-component", "test-action", tt.kv...)
		})
	}
}
