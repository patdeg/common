package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestSecurityHeadersMiddleware verifies that all security headers are set correctly
func TestSecurityHeadersMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantHSTS   bool
		wantCSP    bool
		wantCOOP   bool
		wantCORP   bool
		wantNoSniff bool
		wantCache   bool
	}{
		{
			name:        "API route with cache control",
			path:        "/api/v1/test",
			wantHSTS:    true,
			wantCSP:     true,
			wantCOOP:    true,
			wantCORP:    true,
			wantNoSniff: true,
			wantCache:   true,
		},
		{
			name:        "Auth route with cache control",
			path:        "/auth/login",
			wantHSTS:    true,
			wantCSP:     true,
			wantCOOP:    true,
			wantCORP:    true,
			wantNoSniff: true,
			wantCache:   true,
		},
		{
			name:        "Regular route without cache control",
			path:        "/dashboard",
			wantHSTS:    true,
			wantCSP:     true,
			wantCOOP:    true,
			wantCORP:    true,
			wantNoSniff: true,
			wantCache:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware := SecurityHeadersMiddleware(DefaultSecurityConfig())
			wrapped := middleware(handler)

			req := httptest.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			if tt.wantHSTS {
				hsts := rec.Header().Get("Strict-Transport-Security")
				if hsts == "" {
					t.Error("Expected HSTS header, got none")
				}
				if !strings.Contains(hsts, "max-age=63072000") {
					t.Errorf("Expected HSTS max-age=63072000, got %s", hsts)
				}
			}

			if tt.wantCSP {
				csp := rec.Header().Get("Content-Security-Policy")
				if csp == "" {
					t.Error("Expected CSP header, got none")
				}
				if !strings.Contains(csp, "default-src 'self'") {
					t.Errorf("Expected CSP default-src 'self', got %s", csp)
				}
			}

			if tt.wantCOOP {
				coop := rec.Header().Get("Cross-Origin-Opener-Policy")
				if coop != "same-origin" {
					t.Errorf("Expected COOP same-origin, got %s", coop)
				}
			}

			if tt.wantCORP {
				corp := rec.Header().Get("Cross-Origin-Resource-Policy")
				if corp != "same-origin" {
					t.Errorf("Expected CORP same-origin, got %s", corp)
				}
			}

			if tt.wantNoSniff {
				nosniff := rec.Header().Get("X-Content-Type-Options")
				if nosniff != "nosniff" {
					t.Errorf("Expected X-Content-Type-Options nosniff, got %s", nosniff)
				}
			}

			if tt.wantCache {
				cache := rec.Header().Get("Cache-Control")
				if !strings.Contains(cache, "no-store") {
					t.Errorf("Expected Cache-Control no-store, got %s", cache)
				}
			}

			// Check that dangerous headers are removed
			if xpowered := rec.Header().Get("X-Powered-By"); xpowered != "" {
				t.Errorf("Expected X-Powered-By to be removed, got %s", xpowered)
			}

			// Check X-Frame-Options
			xframe := rec.Header().Get("X-Frame-Options")
			if xframe != "DENY" {
				t.Errorf("Expected X-Frame-Options DENY, got %s", xframe)
			}
		})
	}
}

// TestCORSMiddleware verifies CORS handling for allowed and blocked origins
func TestCORSMiddleware(t *testing.T) {
	tests := []struct {
		name               string
		origin             string
		method             string
		allowedOrigins     []string
		expectAllowed      bool
		expectVaryHeaders  bool
		expectStatusCode   int
	}{
		{
			name:               "Allowed origin",
			origin:             "https://example.com",
			method:             "GET",
			allowedOrigins:     []string{"https://example.com"},
			expectAllowed:      true,
			expectVaryHeaders:  true,
			expectStatusCode:   http.StatusOK,
		},
		{
			name:               "Blocked origin",
			origin:             "https://evil.com",
			method:             "GET",
			allowedOrigins:     []string{"https://example.com"},
			expectAllowed:      false,
			expectVaryHeaders:  false,
			expectStatusCode:   http.StatusOK,
		},
		{
			name:               "Preflight allowed origin",
			origin:             "https://example.com",
			method:             "OPTIONS",
			allowedOrigins:     []string{"https://example.com"},
			expectAllowed:      true,
			expectVaryHeaders:  true,
			expectStatusCode:   http.StatusNoContent,
		},
		{
			name:               "Preflight blocked origin",
			origin:             "https://evil.com",
			method:             "OPTIONS",
			allowedOrigins:     []string{"https://example.com"},
			expectAllowed:      false,
			expectVaryHeaders:  true,
			expectStatusCode:   http.StatusForbidden,
		},
		{
			name:               "Wildcard allowed",
			origin:             "https://anything.com",
			method:             "GET",
			allowedOrigins:     []string{"*"},
			expectAllowed:      true,
			expectVaryHeaders:  true,
			expectStatusCode:   http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			config := DefaultSecurityConfig()
			config.AllowedOrigins = tt.allowedOrigins
			config.AllowedMethods = []string{"GET", "POST", "OPTIONS"}
			config.AllowedHeaders = []string{"Content-Type", "Authorization"}

			middleware := CORSMiddleware(config)
			wrapped := middleware(handler)

			req := httptest.NewRequest(tt.method, "/test", nil)
			req.Header.Set("Origin", tt.origin)
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			if rec.Code != tt.expectStatusCode {
				t.Errorf("Expected status %d, got %d", tt.expectStatusCode, rec.Code)
			}

			acao := rec.Header().Get("Access-Control-Allow-Origin")
			if tt.expectAllowed {
				if acao != tt.origin {
					t.Errorf("Expected Access-Control-Allow-Origin %s, got %s", tt.origin, acao)
				}
			} else {
				if acao != "" && tt.method != "OPTIONS" {
					t.Errorf("Expected no Access-Control-Allow-Origin for blocked origin, got %s", acao)
				}
			}

			if tt.expectVaryHeaders {
				vary := rec.Header().Values("Vary")
				if len(vary) == 0 {
					t.Error("Expected Vary headers, got none")
				}
			}
		})
	}
}

// TestTLSRedirectMiddleware verifies HTTPS redirect behavior
func TestTLSRedirectMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		scheme         string
		xForwardedProto string
		expectRedirect bool
	}{
		{
			name:           "HTTP request should redirect",
			scheme:         "http",
			xForwardedProto: "",
			expectRedirect: true,
		},
		{
			name:           "HTTPS request should not redirect",
			scheme:         "https",
			xForwardedProto: "",
			expectRedirect: false,
		},
		{
			name:           "X-Forwarded-Proto HTTPS should not redirect",
			scheme:         "http",
			xForwardedProto: "https",
			expectRedirect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			wrapped := TLSRedirectMiddleware(handler)

			req := httptest.NewRequest("GET", tt.scheme+"://example.com/test", nil)
			if tt.xForwardedProto != "" {
				req.Header.Set("X-Forwarded-Proto", tt.xForwardedProto)
			}
			rec := httptest.NewRecorder()

			wrapped.ServeHTTP(rec, req)

			if tt.expectRedirect {
				if rec.Code != http.StatusMovedPermanently {
					t.Errorf("Expected redirect status %d, got %d", http.StatusMovedPermanently, rec.Code)
				}
				location := rec.Header().Get("Location")
				if !strings.HasPrefix(location, "https://") {
					t.Errorf("Expected HTTPS redirect, got %s", location)
				}
			} else {
				if rec.Code == http.StatusMovedPermanently {
					t.Error("Did not expect redirect for HTTPS request")
				}
			}
		})
	}
}

// TestSecureCookieConfig verifies cookie security settings
func TestSecureCookieConfig(t *testing.T) {
	tests := []struct {
		name           string
		cookieName     string
		cookiePath     string
		cookieDomain   string
		expectHostPrefix bool
	}{
		{
			name:           "Cookie with __Host- prefix eligible",
			cookieName:     "session",
			cookiePath:     "/",
			cookieDomain:   "",
			expectHostPrefix: true,
		},
		{
			name:           "Cookie with non-root path not eligible",
			cookieName:     "session",
			cookiePath:     "/admin",
			cookieDomain:   "",
			expectHostPrefix: false,
		},
		{
			name:           "Cookie with domain not eligible",
			cookieName:     "session",
			cookiePath:     "/",
			cookieDomain:   "example.com",
			expectHostPrefix: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DefaultSecurityConfig()
			config.CookiePath = tt.cookiePath
			config.CookieDomain = tt.cookieDomain

			cookie := &http.Cookie{
				Name:  tt.cookieName,
				Value: "test-value",
			}

			SecureCookieConfig(cookie, config)

			if !cookie.Secure {
				t.Error("Expected cookie.Secure to be true")
			}

			if !cookie.HttpOnly {
				t.Error("Expected cookie.HttpOnly to be true")
			}

			if cookie.SameSite != http.SameSiteStrictMode {
				t.Errorf("Expected SameSite Strict, got %v", cookie.SameSite)
			}

			if tt.expectHostPrefix {
				if !strings.HasPrefix(cookie.Name, "__Host-") {
					t.Errorf("Expected __Host- prefix, got %s", cookie.Name)
				}
			} else {
				if strings.HasPrefix(cookie.Name, "__Host-") && tt.cookiePath != "/" {
					t.Errorf("Did not expect __Host- prefix for non-root path, got %s", cookie.Name)
				}
			}
		})
	}
}

// TestSanitizeRedirectTarget verifies open redirect protection
func TestSanitizeRedirectTarget(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		def      string
		expected string
	}{
		{
			name:     "Empty string returns default",
			raw:      "",
			def:      "/dashboard",
			expected: "/dashboard",
		},
		{
			name:     "Whitespace only returns default",
			raw:      "   ",
			def:      "/dashboard",
			expected: "/dashboard",
		},
		{
			name:     "Valid relative URL",
			raw:      "/page?foo=bar",
			def:      "/dashboard",
			expected: "/page?foo=bar",
		},
		{
			name:     "Valid relative URL with fragment",
			raw:      "/page#section",
			def:      "/dashboard",
			expected: "/page#section",
		},
		{
			name:     "Absolute URL returns default",
			raw:      "https://evil.com/phishing",
			def:      "/dashboard",
			expected: "/dashboard",
		},
		{
			name:     "Protocol-relative URL returns default",
			raw:      "//evil.com/phishing",
			def:      "/dashboard",
			expected: "/dashboard",
		},
		{
			name:     "Relative path without slash returns default",
			raw:      "page",
			def:      "/dashboard",
			expected: "/dashboard",
		},
		{
			name:     "URL with host returns default",
			raw:      "http://example.com/page",
			def:      "/dashboard",
			expected: "/dashboard",
		},
		{
			name:     "Invalid URL returns default",
			raw:      "ht!tp://invalid",
			def:      "/dashboard",
			expected: "/dashboard",
		},
		{
			name:     "Complex relative URL",
			raw:      "/api/v1/resource?id=123&action=view#top",
			def:      "/dashboard",
			expected: "/api/v1/resource?id=123&action=view#top",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeRedirectTarget(tt.raw, tt.def)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestBuildCSPHeader verifies CSP header construction
func TestBuildCSPHeader(t *testing.T) {
	config := DefaultSecurityConfig()
	csp := buildCSPHeader(config)

	// Check that all expected directives are present
	expectedDirectives := []string{
		"default-src 'self'",
		"script-src",
		"style-src",
		"img-src",
		"font-src",
		"connect-src",
		"frame-src 'none'",
		"object-src 'none'",
		"upgrade-insecure-requests",
	}

	for _, directive := range expectedDirectives {
		if !strings.Contains(csp, directive) {
			t.Errorf("Expected CSP to contain %s, got %s", directive, csp)
		}
	}
}

// TestBuildHSTSHeader verifies HSTS header construction
func TestBuildHSTSHeader(t *testing.T) {
	tests := []struct {
		name               string
		maxAge             int
		includeSubdomains  bool
		preload            bool
		expected           string
	}{
		{
			name:              "Full HSTS header",
			maxAge:            63072000,
			includeSubdomains: true,
			preload:           true,
			expected:          "max-age=63072000; includeSubDomains; preload",
		},
		{
			name:              "HSTS without preload",
			maxAge:            31536000,
			includeSubdomains: true,
			preload:           false,
			expected:          "max-age=31536000; includeSubDomains",
		},
		{
			name:              "HSTS without subdomains",
			maxAge:            31536000,
			includeSubdomains: false,
			preload:           false,
			expected:          "max-age=31536000",
		},
		{
			name:              "HSTS disabled",
			maxAge:            0,
			includeSubdomains: false,
			preload:           false,
			expected:          "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &SecurityConfig{
				HSTSMaxAge:            tt.maxAge,
				HSTSIncludeSubdomains: tt.includeSubdomains,
				HSTSPreload:           tt.preload,
			}

			result := buildHSTSHeader(config)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestBuildPermissionsPolicyHeader verifies Permissions-Policy header construction
func TestBuildPermissionsPolicyHeader(t *testing.T) {
	config := DefaultSecurityConfig()
	policy := buildPermissionsPolicyHeader(config)

	// Check that some expected permissions are present
	expectedPermissions := []string{
		"geolocation=()",
		"camera=()",
		"microphone=()",
	}

	for _, perm := range expectedPermissions {
		if !strings.Contains(policy, perm) {
			t.Errorf("Expected Permissions-Policy to contain %s, got %s", perm, policy)
		}
	}
}

// TestRateLimitHeaders verifies rate limit header formatting
func TestRateLimitHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	RateLimitHeaders(rec, 1000, 500, 1234567890)

	limit := rec.Header().Get("X-RateLimit-Limit")
	if limit != "1000" {
		t.Errorf("Expected X-RateLimit-Limit 1000, got %s", limit)
	}

	remaining := rec.Header().Get("X-RateLimit-Remaining")
	if remaining != "500" {
		t.Errorf("Expected X-RateLimit-Remaining 500, got %s", remaining)
	}

	reset := rec.Header().Get("X-RateLimit-Reset")
	if reset != "1234567890" {
		t.Errorf("Expected X-RateLimit-Reset 1234567890, got %s", reset)
	}
}
