// Copyright 2025 Patrick Deglon
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package csrf provides Cross-Site Request Forgery (CSRF) protection middleware
// for HTTP handlers. It generates cryptographically secure tokens for safe methods
// (GET, HEAD, OPTIONS) and validates them for state-changing methods (POST, PUT,
// DELETE, PATCH).
//
// Usage:
//
//	store := csrf.NewTokenStore()
//	handler := store.Middleware(yourHandler)
//	http.ListenAndServe(":8080", handler)
//
// In HTML forms, include the CSRF token as a hidden field:
//
//	<form method="POST">
//	    <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
//	    <!-- other fields -->
//	</form>
//
// For AJAX requests, include the token in the X-CSRF-Token header:
//
//	fetch('/api/endpoint', {
//	    method: 'POST',
//	    headers: {
//	        'X-CSRF-Token': getCookieValue('csrf_token')
//	    }
//	})
package csrf

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	tokenLength = 32
	cookieName  = "csrf_token"
	headerName  = "X-CSRF-Token"
	formField   = "csrf_token"
)

// TokenStore manages CSRF tokens with automatic expiry and cleanup
type TokenStore struct {
	mu     sync.RWMutex
	tokens map[string]time.Time
}

// NewTokenStore creates a new token store and starts a background cleanup goroutine
// that removes expired tokens every hour
func NewTokenStore() *TokenStore {
	store := &TokenStore{
		tokens: make(map[string]time.Time),
	}
	// Cleanup expired tokens periodically
	go store.cleanup()
	return store
}

// GenerateToken creates a cryptographically secure random token
// Returns the base64-encoded token string or an error if random generation fails
func (ts *TokenStore) GenerateToken() (string, error) {
	b := make([]byte, tokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate CSRF token: %w", err)
	}

	token := base64.URLEncoding.EncodeToString(b)

	// Store token with 24-hour expiry
	ts.mu.Lock()
	ts.tokens[token] = time.Now().Add(24 * time.Hour)
	ts.mu.Unlock()

	return token, nil
}

// ValidateToken checks if a token is valid and not expired
// Returns true if the token exists and has not expired
func (ts *TokenStore) ValidateToken(token string) bool {
	ts.mu.RLock()
	expiry, exists := ts.tokens[token]
	ts.mu.RUnlock()

	if !exists {
		return false
	}

	if time.Now().After(expiry) {
		// Token expired, remove it
		ts.mu.Lock()
		delete(ts.tokens, token)
		ts.mu.Unlock()
		return false
	}

	return true
}

// cleanup removes expired tokens from the store
// Runs continuously in a background goroutine
func (ts *TokenStore) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		ts.mu.Lock()
		now := time.Now()
		for token, expiry := range ts.tokens {
			if now.After(expiry) {
				delete(ts.tokens, token)
			}
		}
		ts.mu.Unlock()
	}
}

// Middleware provides CSRF protection for HTTP handlers
// Safe methods (GET, HEAD, OPTIONS) generate and set a new token
// State-changing methods (POST, PUT, DELETE, PATCH) validate the token
func (ts *TokenStore) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate and set token for safe methods
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			token, err := ts.GenerateToken()
			if err != nil {
				http.Error(w, "Failed to generate CSRF token", http.StatusInternalServerError)
				return
			}

			// Determine if connection is secure
			isSecure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
			if r.Host == "localhost" || r.Host == "127.0.0.1" {
				isSecure = false // Allow insecure cookies on localhost for development
			}

			// Set cookie (HttpOnly=false so JavaScript can read it for AJAX)
			http.SetCookie(w, &http.Cookie{
				Name:     cookieName,
				Value:    token,
				Path:     "/",
				HttpOnly: false, // JavaScript needs to read this for AJAX requests
				Secure:   isSecure,
				SameSite: http.SameSiteStrictMode,
				MaxAge:   86400, // 24 hours
			})

			next.ServeHTTP(w, r)
			return
		}

		// Validate token for state-changing methods
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE" || r.Method == "PATCH" {
			cookieToken, err := r.Cookie(cookieName)
			if err != nil {
				http.Error(w, "CSRF token cookie missing", http.StatusForbidden)
				return
			}

			// Check header first (for AJAX), then form field
			requestToken := r.Header.Get(headerName)
			if requestToken == "" {
				// Parse form to get csrf_token field
				if err := r.ParseForm(); err == nil {
					requestToken = r.FormValue(formField)
				}
			}

			if requestToken == "" {
				http.Error(w, "CSRF token missing from request", http.StatusForbidden)
				return
			}

			// Validate token exists and hasn't expired
			if !ts.ValidateToken(requestToken) {
				http.Error(w, "CSRF token invalid or expired", http.StatusForbidden)
				return
			}

			// Constant-time comparison to prevent timing attacks
			if subtle.ConstantTimeCompare([]byte(cookieToken.Value), []byte(requestToken)) != 1 {
				http.Error(w, "CSRF token validation failed", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
			return
		}

		// Other methods (TRACE, CONNECT, etc.) - pass through
		next.ServeHTTP(w, r)
	})
}

// GetToken retrieves the CSRF token from the request cookie
// This helper function is useful for injecting the token into templates
// Returns empty string if the cookie is not present
func GetToken(r *http.Request) string {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}
