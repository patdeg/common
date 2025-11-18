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

package csrf

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestTokenGeneration(t *testing.T) {
	store := NewTokenStore()

	t.Run("generates unique tokens", func(t *testing.T) {
		token1, err := store.GenerateToken()
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		token2, err := store.GenerateToken()
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		if token1 == token2 {
			t.Error("Generated tokens should be unique")
		}

		if len(token1) == 0 {
			t.Error("Token should not be empty")
		}
	})

	t.Run("validates newly generated token", func(t *testing.T) {
		token, err := store.GenerateToken()
		if err != nil {
			t.Fatalf("Failed to generate token: %v", err)
		}

		if !store.ValidateToken(token) {
			t.Error("Newly generated token should be valid")
		}
	})

	t.Run("rejects invalid token", func(t *testing.T) {
		if store.ValidateToken("invalid-token-xyz") {
			t.Error("Invalid token should not validate")
		}
	})
}

func TestCSRFMiddleware(t *testing.T) {
	store := NewTokenStore()

	// Test handler that just returns 200 OK
	handler := store.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	t.Run("GET request generates token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}

		cookies := w.Result().Cookies()
		if len(cookies) == 0 {
			t.Fatal("No CSRF cookie set")
		}

		var csrfCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == cookieName {
				csrfCookie = c
				break
			}
		}

		if csrfCookie == nil {
			t.Fatal("CSRF cookie not found")
		}

		if csrfCookie.Value == "" {
			t.Error("CSRF token is empty")
		}

		if csrfCookie.MaxAge != 86400 {
			t.Errorf("Expected MaxAge 86400, got %d", csrfCookie.MaxAge)
		}

		if csrfCookie.SameSite != http.SameSiteStrictMode {
			t.Error("Expected SameSite=Strict")
		}
	})

	t.Run("POST without token fails", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected 403, got %d", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "CSRF token") {
			t.Errorf("Error message should mention CSRF token, got: %s", body)
		}
	})

	t.Run("POST with valid token in form succeeds", func(t *testing.T) {
		// First, get a token
		getReq := httptest.NewRequest("GET", "/", nil)
		getW := httptest.NewRecorder()
		handler.ServeHTTP(getW, getReq)

		var token string
		for _, c := range getW.Result().Cookies() {
			if c.Name == cookieName {
				token = c.Value
				break
			}
		}

		// Now POST with the token in form
		form := url.Values{}
		form.Set(formField, token)
		postReq := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
		postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		postReq.AddCookie(&http.Cookie{Name: cookieName, Value: token})

		postW := httptest.NewRecorder()
		handler.ServeHTTP(postW, postReq)

		if postW.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d: %s", postW.Code, postW.Body.String())
		}

		if postW.Body.String() != "success" {
			t.Errorf("Expected 'success', got %s", postW.Body.String())
		}
	})

	t.Run("POST with valid token in header succeeds", func(t *testing.T) {
		// Get token
		getReq := httptest.NewRequest("GET", "/", nil)
		getW := httptest.NewRecorder()
		handler.ServeHTTP(getW, getReq)

		var token string
		for _, c := range getW.Result().Cookies() {
			if c.Name == cookieName {
				token = c.Value
				break
			}
		}

		// POST with header
		postReq := httptest.NewRequest("POST", "/", nil)
		postReq.Header.Set(headerName, token)
		postReq.AddCookie(&http.Cookie{Name: cookieName, Value: token})

		postW := httptest.NewRecorder()
		handler.ServeHTTP(postW, postReq)

		if postW.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d: %s", postW.Code, postW.Body.String())
		}
	})

	t.Run("POST with mismatched token fails", func(t *testing.T) {
		token1, _ := store.GenerateToken()
		token2, _ := store.GenerateToken()

		form := url.Values{}
		form.Set(formField, token2)
		req := httptest.NewRequest("POST", "/", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: cookieName, Value: token1})

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected 403, got %d", w.Code)
		}
	})

	t.Run("PUT request validates token", func(t *testing.T) {
		// GET token
		getReq := httptest.NewRequest("GET", "/", nil)
		getW := httptest.NewRecorder()
		handler.ServeHTTP(getW, getReq)

		var token string
		for _, c := range getW.Result().Cookies() {
			if c.Name == cookieName {
				token = c.Value
				break
			}
		}

		// PUT with token
		putReq := httptest.NewRequest("PUT", "/", nil)
		putReq.Header.Set(headerName, token)
		putReq.AddCookie(&http.Cookie{Name: cookieName, Value: token})

		putW := httptest.NewRecorder()
		handler.ServeHTTP(putW, putReq)

		if putW.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", putW.Code)
		}
	})

	t.Run("DELETE request validates token", func(t *testing.T) {
		// GET token
		getReq := httptest.NewRequest("GET", "/", nil)
		getW := httptest.NewRecorder()
		handler.ServeHTTP(getW, getReq)

		var token string
		for _, c := range getW.Result().Cookies() {
			if c.Name == cookieName {
				token = c.Value
				break
			}
		}

		// DELETE with token
		deleteReq := httptest.NewRequest("DELETE", "/", nil)
		deleteReq.Header.Set(headerName, token)
		deleteReq.AddCookie(&http.Cookie{Name: cookieName, Value: token})

		deleteW := httptest.NewRecorder()
		handler.ServeHTTP(deleteW, deleteReq)

		if deleteW.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", deleteW.Code)
		}
	})

	t.Run("PATCH request validates token", func(t *testing.T) {
		// GET token
		getReq := httptest.NewRequest("GET", "/", nil)
		getW := httptest.NewRecorder()
		handler.ServeHTTP(getW, getReq)

		var token string
		for _, c := range getW.Result().Cookies() {
			if c.Name == cookieName {
				token = c.Value
				break
			}
		}

		// PATCH with token
		patchReq := httptest.NewRequest("PATCH", "/", nil)
		patchReq.Header.Set(headerName, token)
		patchReq.AddCookie(&http.Cookie{Name: cookieName, Value: token})

		patchW := httptest.NewRecorder()
		handler.ServeHTTP(patchW, patchReq)

		if patchW.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", patchW.Code)
		}
	})

	t.Run("HEAD request generates token", func(t *testing.T) {
		req := httptest.NewRequest("HEAD", "/", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}

		var foundToken bool
		for _, c := range w.Result().Cookies() {
			if c.Name == cookieName {
				foundToken = true
				break
			}
		}

		if !foundToken {
			t.Error("HEAD request should generate CSRF token")
		}
	})

	t.Run("OPTIONS request generates token", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}

		var foundToken bool
		for _, c := range w.Result().Cookies() {
			if c.Name == cookieName {
				foundToken = true
				break
			}
		}

		if !foundToken {
			t.Error("OPTIONS request should generate CSRF token")
		}
	})
}

func TestTokenExpiry(t *testing.T) {
	store := NewTokenStore()

	token, err := store.GenerateToken()
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Token should be valid immediately
	if !store.ValidateToken(token) {
		t.Error("Token should be valid immediately after generation")
	}

	// Manually expire the token by setting it to the past
	store.mu.Lock()
	store.tokens[token] = time.Now().Add(-1 * time.Hour)
	store.mu.Unlock()

	// Token should now be invalid
	if store.ValidateToken(token) {
		t.Error("Expired token should not validate")
	}

	// Token should be removed from store after validation fails
	store.mu.RLock()
	_, exists := store.tokens[token]
	store.mu.RUnlock()

	if exists {
		t.Error("Expired token should be removed from store")
	}
}

func TestGetToken(t *testing.T) {
	t.Run("returns token from cookie", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: cookieName, Value: "test-token-123"})

		token := GetToken(req)
		if token != "test-token-123" {
			t.Errorf("Expected 'test-token-123', got '%s'", token)
		}
	})

	t.Run("returns empty string when cookie missing", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)

		token := GetToken(req)
		if token != "" {
			t.Errorf("Expected empty string, got '%s'", token)
		}
	})
}

func TestSecureCookie(t *testing.T) {
	store := NewTokenStore()
	handler := store.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("sets Secure=true for HTTPS", func(t *testing.T) {
		req := httptest.NewRequest("GET", "https://example.com/", nil)
		req.Header.Set("X-Forwarded-Proto", "https")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		var csrfCookie *http.Cookie
		for _, c := range w.Result().Cookies() {
			if c.Name == cookieName {
				csrfCookie = c
				break
			}
		}

		if csrfCookie == nil {
			t.Fatal("CSRF cookie not found")
		}

		if !csrfCookie.Secure {
			t.Error("Cookie should be Secure for HTTPS connections")
		}
	})

	t.Run("sets Secure=false for localhost", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://localhost/", nil)
		req.Host = "localhost"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		var csrfCookie *http.Cookie
		for _, c := range w.Result().Cookies() {
			if c.Name == cookieName {
				csrfCookie = c
				break
			}
		}

		if csrfCookie == nil {
			t.Fatal("CSRF cookie not found")
		}

		if csrfCookie.Secure {
			t.Error("Cookie should not be Secure for localhost")
		}
	})
}

func BenchmarkGenerateToken(b *testing.B) {
	store := NewTokenStore()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := store.GenerateToken()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidateToken(b *testing.B) {
	store := NewTokenStore()
	token, _ := store.GenerateToken()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		store.ValidateToken(token)
	}
}

func BenchmarkMiddleware(b *testing.B) {
	store := NewTokenStore()
	handler := store.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}
