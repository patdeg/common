package web

// This file centralizes HTTP security concerns for web applications.
//
// Design goals and guardrails:
// - Provide strong defaults that are safe for production traffic.
// - Keep configuration explicit; "deny by default" for CORS and CSP sources.
// - Prefer standard headers (CSP, HSTS, Permissions-Policy) over ad-hoc checks.
// - Be proxy/CDN friendly: set Vary headers where responses depend on request headers.
// - Avoid brittle logic that might break behind AppEngine/Cloud Load Balancers.
// - Document intent inline to make trade-offs clear for future maintainers.
//
// Important security notes:
// - CSP defaults include 'unsafe-inline' for script/style to support the existing
//   SSR + Bootstrap + HTMX stack. For stricter deployments, prefer nonces/hashes
//   and remove 'unsafe-inline'. This should be changed alongside template updates.
// - CORS echoes the request Origin only when explicitly allowed, never "*" when
//   credentials are enabled. We also add appropriate Vary headers to prevent
//   cache poisoning or cross-tenant cache hits.
// - HSTS is enabled by default with preload-friendly settings; ensure HTTPS is
//   consistently used before enabling preload in production environments.
// - Cookie helpers apply Secure, HttpOnly, SameSite, and opt into __Host- prefix
//   when safe, minimizing cookie scope and mitigating cookie injection risks.

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	// CSP configuration
	CSPDefaultSrc     []string
	CSPScriptSrc      []string
	CSPStyleSrc       []string
	CSPImgSrc         []string
	CSPFontSrc        []string
	CSPConnectSrc     []string
	CSPFrameSrc       []string
	CSPObjectSrc      []string
	CSPMediaSrc       []string
	CSPWorkerSrc      []string
	CSPManifestSrc    []string
	CSPFormAction     []string
	CSPFrameAncestors []string
	CSPBaseURI        []string

	// CORS configuration
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	MaxAge           int
	AllowCredentials bool

	// HSTS configuration
	HSTSMaxAge            int
	HSTSIncludeSubdomains bool
	HSTSPreload           bool

	// Cookie configuration
	CookieDomain   string
	CookiePath     string
	CookieSecure   bool
	CookieHTTPOnly bool
	CookieSameSite http.SameSite

	// Feature policy / Permissions policy
	PermissionsPolicy map[string]string
}

// DefaultSecurityConfig returns the default security configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		// CSP - restrictive by default
		CSPDefaultSrc: []string{"'self'"},
		// Note: 'unsafe-inline' is retained to avoid breaking existing SSR
		// templates and Bootstrap usage. Prefer nonces/hashes and remove this
		// in hardened deployments.
		CSPScriptSrc:      []string{"'self'", "'unsafe-inline'", "https://unpkg.com", "https://cdn.jsdelivr.net", "https://cdnjs.cloudflare.com"}, // HTMX, Bootstrap, Chart.js, html2canvas
		CSPStyleSrc:       []string{"'self'", "'unsafe-inline'", "https://cdn.jsdelivr.net", "https://fonts.googleapis.com"},                      // Bootstrap, Google Fonts
		CSPImgSrc:         []string{"'self'", "data:", "https:"},
		CSPFontSrc:        []string{"'self'", "data:", "https://cdn.jsdelivr.net", "https://fonts.gstatic.com"},
		CSPConnectSrc:     []string{"'self'", "https://cdn.jsdelivr.net", "https://unpkg.com"}, // Allow CDN source maps
		CSPFrameSrc:       []string{"'none'"},
		CSPObjectSrc:      []string{"'none'"},
		CSPMediaSrc:       []string{"'none'"},
		CSPWorkerSrc:      []string{"'none'"},
		CSPManifestSrc:    []string{"'self'"},
		CSPFormAction:     []string{"'self'"},
		CSPFrameAncestors: []string{"'none'"},
		CSPBaseURI:        []string{"'self'"},

		// CORS - deny by default
		AllowedOrigins:   []string{},
		AllowedMethods:   []string{},
		AllowedHeaders:   []string{},
		MaxAge:           3600,
		AllowCredentials: false,

		// HSTS - strict configuration
		HSTSMaxAge:            63072000, // 2 years
		HSTSIncludeSubdomains: true,
		HSTSPreload:           true,

		// Cookie security
		CookieDomain:   "",
		CookiePath:     "/",
		CookieSecure:   true,
		CookieHTTPOnly: true,
		CookieSameSite: http.SameSiteStrictMode,

		// Permissions Policy - restrictive by default
		PermissionsPolicy: map[string]string{
			"geolocation":        "()",
			"microphone":         "()",
			"camera":             "()",
			"payment":            "()",
			"usb":                "()",
			"magnetometer":       "()",
			"gyroscope":          "()",
			"accelerometer":      "()",
			"autoplay":           "()",
			"encrypted-media":    "()",
			"picture-in-picture": "()",
			"fullscreen":         "(self)",
			"sync-xhr":           "()",
		},
	}
}

// SecurityHeadersMiddleware adds comprehensive security headers to all responses
func SecurityHeadersMiddleware(config *SecurityConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	// Pre-build static headers for performance
	cspHeader := buildCSPHeader(config)
	hstsHeader := buildHSTSHeader(config)
	permissionsPolicyHeader := buildPermissionsPolicyHeader(config)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// All headers below are set prior to calling the next handler so
			// downstream handlers don't need to worry about duplicating logic.
			// If a handler needs to override a header, it may do so explicitly.
			// HTTPS enforcement via HSTS
			if hstsHeader != "" {
				w.Header().Set("Strict-Transport-Security", hstsHeader)
			}

			// Content Security Policy
			if cspHeader != "" {
				w.Header().Set("Content-Security-Policy", cspHeader)
			}

			// Cross-origin isolation defaults (defense-in-depth):
			// - COOP prevents the page from being put in a browsing context
			//   group with cross-origin documents, mitigating certain XS-Leaks.
			// - CORP prevents cross-origin documents from loading this resource
			//   unless explicitly allowed by the other origin.
			// These are safe for our server-rendered UI and API responses.
			w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
			w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")

			// X-Frame-Options (legacy, but still useful)
			w.Header().Set("X-Frame-Options", "DENY")

			// X-Content-Type-Options
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// X-XSS-Protection (legacy, but harmless)
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Referrer-Policy
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Permissions-Policy (formerly Feature-Policy)
			if permissionsPolicyHeader != "" {
				w.Header().Set("Permissions-Policy", permissionsPolicyHeader)
			}

			// Remove potentially dangerous headers
			w.Header().Del("X-Powered-By")
			w.Header().Del("Server")

			// Cache-Control for sensitive content
			if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/auth/") {
				w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
				w.Header().Set("Pragma", "no-cache")
				w.Header().Set("Expires", "0")
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing with a blocklist approach
func CORSMiddleware(config *SecurityConfig) func(http.Handler) http.Handler {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Default: deny all CORS requests
			corsAllowed := false

			// Check if origin is in allowed list
			if origin != "" && len(config.AllowedOrigins) > 0 {
				for _, allowed := range config.AllowedOrigins {
					if allowed == "*" || allowed == origin {
						corsAllowed = true
						break
					}
				}
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				if corsAllowed {
					// Set CORS headers for allowed origins
					w.Header().Set("Access-Control-Allow-Origin", origin)
					// Inform caches that response varies by Origin and preflight headers
					w.Header().Add("Vary", "Origin")
					w.Header().Add("Vary", "Access-Control-Request-Method")
					w.Header().Add("Vary", "Access-Control-Request-Headers")

					if len(config.AllowedMethods) > 0 {
						w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
					}

					if len(config.AllowedHeaders) > 0 {
						w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
					}

					if config.MaxAge > 0 {
						// Use proper integer-to-string conversion for header values
						w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
					}

					if config.AllowCredentials {
						w.Header().Set("Access-Control-Allow-Credentials", "true")
					}

					w.WriteHeader(http.StatusNoContent)
					return
				}

				// Deny preflight for non-allowed origins
				// Also add Vary so negative decisions are not cached broadly.
				w.Header().Add("Vary", "Origin")
				w.Header().Add("Vary", "Access-Control-Request-Method")
				w.Header().Add("Vary", "Access-Control-Request-Headers")
				w.WriteHeader(http.StatusForbidden)
				return
			}

			// Set CORS headers for actual requests from allowed origins
			if corsAllowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				// Ensure caches keep per-origin variants for resource responses
				w.Header().Add("Vary", "Origin")

				if config.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}

				// Expose certain headers to the client
				w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID, X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset")
			}

			// Continue processing
			next.ServeHTTP(w, r)
		})
	}
}

// TLSRedirectMiddleware redirects HTTP requests to HTTPS
func TLSRedirectMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request is already HTTPS
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			next.ServeHTTP(w, r)
			return
		}

		// Build HTTPS URL
		httpsURL := "https://" + r.Host + r.RequestURI

		// Permanent redirect to HTTPS
		http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
	})
}

// SecureCookieConfig applies secure cookie settings to a cookie
func SecureCookieConfig(cookie *http.Cookie, config *SecurityConfig) {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	// Apply security settings
	cookie.Secure = config.CookieSecure
	cookie.HttpOnly = config.CookieHTTPOnly
	cookie.SameSite = config.CookieSameSite

	if config.CookiePath != "" {
		cookie.Path = config.CookiePath
	}

	if config.CookieDomain != "" {
		cookie.Domain = config.CookieDomain
	}

	// Use __Host- prefix for maximum security when possible
	if cookie.Secure && cookie.Path == "/" && cookie.Domain == "" {
		if !strings.HasPrefix(cookie.Name, "__Host-") {
			cookie.Name = "__Host-" + cookie.Name
		}
	}
}

// SanitizeRedirectTarget validates and sanitizes redirect URLs to prevent open redirects.
// It returns the sanitized relative URL if safe, otherwise returns the default URL.
// Only allows relative URLs starting with "/" on the same host.
func SanitizeRedirectTarget(raw string, def string) string {
	// Fast path: empty falls back to default.
	if strings.TrimSpace(raw) == "" {
		return def
	}

	// Parse to inspect components; invalid parse falls back to default.
	u, err := url.Parse(raw)
	if err != nil {
		return def
	}

	// Disallow absolute URLs and protocol-relative redirects.
	if u.IsAbs() || u.Host != "" || strings.HasPrefix(raw, "//") {
		return def
	}

	// Require an absolute path that starts with "/".
	if u.Path == "" || !strings.HasPrefix(u.Path, "/") {
		return def
	}

	// Reconstruct the safe relative target including query/fragment.
	target := u.Path
	if u.RawQuery != "" {
		target += "?" + u.RawQuery
	}
	if u.Fragment != "" {
		target += "#" + u.Fragment
	}
	return target
}

// buildCSPHeader constructs the Content Security Policy header
func buildCSPHeader(config *SecurityConfig) string {
	var directives []string

	if len(config.CSPDefaultSrc) > 0 {
		// Establishes baseline for all fetch directives; typically only 'self'
		// to restrict to same-origin unless otherwise allowed below.
		directives = append(directives, "default-src "+strings.Join(config.CSPDefaultSrc, " "))
	}

	if len(config.CSPScriptSrc) > 0 {
		// Controls JavaScript sources. Consider replacing 'unsafe-inline' with
		// nonces/hashes in hardened deployments to block inline scripts.
		directives = append(directives, "script-src "+strings.Join(config.CSPScriptSrc, " "))
	}

	if len(config.CSPStyleSrc) > 0 {
		// Controls CSS sources. Inline styles are allowed by default to align
		// with Bootstrap SSR usage; prefer hashes in stricter environments.
		directives = append(directives, "style-src "+strings.Join(config.CSPStyleSrc, " "))
	}

	if len(config.CSPImgSrc) > 0 {
		// Permit images from same-origin plus data: URIs if needed for icons.
		directives = append(directives, "img-src "+strings.Join(config.CSPImgSrc, " "))
	}

	if len(config.CSPFontSrc) > 0 {
		// Restrict font sources to trusted CDNs and same-origin.
		directives = append(directives, "font-src "+strings.Join(config.CSPFontSrc, " "))
	}

	if len(config.CSPConnectSrc) > 0 {
		// Controls XHR/fetch/WebSocket connect destinations; typically 'self'.
		directives = append(directives, "connect-src "+strings.Join(config.CSPConnectSrc, " "))
	}

	if len(config.CSPFrameSrc) > 0 {
		// Controls which origins can embed frames from this page; 'none' by default.
		directives = append(directives, "frame-src "+strings.Join(config.CSPFrameSrc, " "))
	}

	if len(config.CSPObjectSrc) > 0 {
		// Legacy plugin content; should be 'none' to disable Flash/Silverlight.
		directives = append(directives, "object-src "+strings.Join(config.CSPObjectSrc, " "))
	}

	if len(config.CSPMediaSrc) > 0 {
		// Audio/video sources; locked down by default.
		directives = append(directives, "media-src "+strings.Join(config.CSPMediaSrc, " "))
	}

	if len(config.CSPWorkerSrc) > 0 {
		// Web worker import sources; disabled by default.
		directives = append(directives, "worker-src "+strings.Join(config.CSPWorkerSrc, " "))
	}

	if len(config.CSPManifestSrc) > 0 {
		// PWA manifest sources; typically 'self'.
		directives = append(directives, "manifest-src "+strings.Join(config.CSPManifestSrc, " "))
	}

	if len(config.CSPFormAction) > 0 {
		// Controls where forms can POST to; usually limited to 'self'.
		directives = append(directives, "form-action "+strings.Join(config.CSPFormAction, " "))
	}

	if len(config.CSPFrameAncestors) > 0 {
		// Prevent clickjacking by restricting which sites may frame this site.
		directives = append(directives, "frame-ancestors "+strings.Join(config.CSPFrameAncestors, " "))
	}

	if len(config.CSPBaseURI) > 0 {
		// Restrict where <base> can point to; typically 'self'.
		directives = append(directives, "base-uri "+strings.Join(config.CSPBaseURI, " "))
	}

	// Always upgrade insecure requests
	directives = append(directives, "upgrade-insecure-requests")

	return strings.Join(directives, "; ")
}

// buildHSTSHeader constructs the HSTS header
func buildHSTSHeader(config *SecurityConfig) string {
	if config.HSTSMaxAge <= 0 {
		return ""
	}

	// Convert numeric max-age using strconv to avoid invalid single-rune strings
	header := "max-age=" + strconv.Itoa(config.HSTSMaxAge)

	if config.HSTSIncludeSubdomains {
		header += "; includeSubDomains"
	}

	if config.HSTSPreload {
		header += "; preload"
	}

	return header
}

// buildPermissionsPolicyHeader constructs the Permissions-Policy header
func buildPermissionsPolicyHeader(config *SecurityConfig) string {
	if len(config.PermissionsPolicy) == 0 {
		return ""
	}

	var policies []string
	for feature, policy := range config.PermissionsPolicy {
		// Map key-value pairs into the header format: feature=policy
		policies = append(policies, feature+"="+policy)
	}

	return strings.Join(policies, ", ")
}

// RateLimitHeaders adds rate limit information to response headers
func RateLimitHeaders(w http.ResponseWriter, limit int, remaining int, resetTime int64) {
	// Properly encode numeric values for HTTP headers
	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetTime, 10))
}
