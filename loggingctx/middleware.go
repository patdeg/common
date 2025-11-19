package loggingctx

import (
	"net/http"
	"time"

	"github.com/patdeg/common"
)

const (
	// requestIDHeader is the HTTP header name for request IDs.
	requestIDHeader = "X-Request-ID"
)

// RequestIDMiddleware is HTTP middleware that generates a unique request ID for each request
// and stores it in the request context. The request ID is also added to the response headers
// for client-side correlation.
//
// Request ID format: ULID-like (26 characters, timestamp-prefixed, URL-safe)
// Example: 01HQZX9P7J8K5M6N4Q3R2T1W0V
//
// If the request already has an X-Request-ID header, it will be preserved.
// Otherwise, a new ULID is generated.
//
// Usage:
//
//	mux := http.NewServeMux()
//	// ... register handlers ...
//	handler := loggingctx.RequestIDMiddleware(mux)
//	http.ListenAndServe(":8080", handler)
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request already has a request ID
		requestID := r.Header.Get(requestIDHeader)
		if requestID == "" {
			// Generate new ULID-like request ID
			requestID = generateULID()
		}

		// Store request ID in context
		ctx := WithRequestID(r.Context(), requestID)
		r = r.WithContext(ctx)

		// Add request ID to response headers for client-side correlation
		w.Header().Set(requestIDHeader, requestID)

		// Call next handler
		next.ServeHTTP(w, r)
	})
}

// LogRequest logs a standardized HTTP request with key metrics.
// Extracts request ID from context for distributed tracing.
//
// Parameters:
//   - ctx: Request context (should contain request ID from RequestIDMiddleware)
//   - r: HTTP request
//   - status: HTTP status code
//   - latency: Request processing duration
//   - bytes: Response size in bytes
func LogRequest(ctx *http.Request, status int, latency time.Duration, bytes int) {
	// Extract request ID from context (set by RequestIDMiddleware)
	reqID := GetRequestID(ctx.Context())
	if reqID == "" {
		// Fallback to header if not in context
		reqID = ctx.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = ctx.Header.Get("X-Request-Id")
		}
	}

	ua := ctx.Header.Get("User-Agent")
	// Keep it short to avoid noisy logs
	if len(ua) > 120 {
		ua = ua[:117] + "..."
	}

	// Log at info level; error/warn severity handled separately by caller
	common.Info("request: method=%s path=%s status=%d latency_ms=%d bytes=%d req_id=%s remote=%s ua=\"%s\"",
		ctx.Method,
		ctx.URL.Path,
		status,
		latency.Milliseconds(),
		bytes,
		reqID,
		ctx.RemoteAddr,
		ua,
	)
}

// LogCall logs a standardized line for handler/API invocations with critical details.
// Always logs at INFO; include additional method/path/query and req_id. Extra kv pairs are appended.
//
// Parameters:
//   - r: HTTP request
//   - component: Component name (e.g., "api", "web", "worker")
//   - action: Action name (e.g., "GetUser", "CreateInteraction")
//   - kv: Optional key-value pairs for additional context (must be even number of arguments)
func LogCall(r *http.Request, component, action string, kv ...interface{}) {
	reqID := r.Header.Get("X-Request-ID")
	if reqID == "" {
		reqID = r.Header.Get("X-Request-Id")
	}
	// Build extra key=value string from kv pairs
	// Build format and args list dynamically
	fmtStr := "call: component=%s action=%s method=%s path=%s query=\"%s\" req_id=%s"
	args := []interface{}{component, action, r.Method, r.URL.Path, r.URL.RawQuery, reqID}
	if len(kv) > 0 {
		for i := 0; i+1 < len(kv); i += 2 {
			k, ok := kv[i].(string)
			if !ok {
				continue
			}
			fmtStr += " " + k + "=%v"
			args = append(args, kv[i+1])
		}
	}
	common.Info(fmtStr, args...)
}
