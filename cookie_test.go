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

// Package common contains tests for cookie helpers.
package common

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestGetCookieIDProductionHost verifies that GetCookieID sets the expected
// security attributes and domain when running on a production host.
func TestGetCookieIDProductionHost(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://example.com/", nil)
	r.Host = "example.com:8080"
	id := GetCookieID(w, r)
	if id == "" {
		t.Fatal("empty id")
	}
	res := w.Result()
	defer res.Body.Close()
	var c *http.Cookie
	for _, cookie := range res.Cookies() {
		if cookie.Name == "ID" {
			c = cookie
			break
		}
	}
	if c == nil {
		t.Fatal("Set-Cookie header not found")
	}
	if !c.HttpOnly {
		t.Error("HttpOnly not set")
	}
	if !c.Secure {
		t.Error("Secure not set")
	}
	if c.Path != "/" {
		t.Errorf("Path = %q, want /", c.Path)
	}
	if !c.Expires.After(time.Now().Add(29*24*time.Hour)) ||
		!c.Expires.Before(time.Now().Add(31*24*time.Hour)) {
		t.Errorf("Expires = %v, want about 30 days", c.Expires)
	}
	if c.Domain != "example.com" {
		t.Errorf("Domain = %q, want example.com", c.Domain)
	}
}

// TestGetCookieIDLocalhost verifies that the domain attribute is omitted when
// running on localhost.
func TestGetCookieIDLocalhost(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://localhost/", nil)
	r.Host = "localhost:8080"
	_ = GetCookieID(w, r)
	res := w.Result()
	defer res.Body.Close()
	var c *http.Cookie
	for _, cookie := range res.Cookies() {
		if cookie.Name == "ID" {
			c = cookie
			break
		}
	}
	if c == nil {
		t.Fatal("Set-Cookie header not found")
	}
	if c.Domain != "" {
		t.Errorf("Domain = %q, want empty", c.Domain)
	}
}

// TestGetCookieIDSecureFlagLocalhost verifies that the Secure flag is false
// when running on localhost to support local development.
func TestGetCookieIDSecureFlagLocalhost(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://localhost/", nil)
	r.Host = "localhost:8080"
	_ = GetCookieID(w, r)
	res := w.Result()
	defer res.Body.Close()
	var c *http.Cookie
	for _, cookie := range res.Cookies() {
		if cookie.Name == "ID" {
			c = cookie
			break
		}
	}
	if c == nil {
		t.Fatal("Set-Cookie header not found")
	}
	if c.Secure {
		t.Error("Secure flag should be false on localhost")
	}
}

// TestGetCookieIDLocalIP verifies that the domain attribute is omitted when the
// host is an IP address on localhost.
func TestGetCookieIDLocalIP(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://127.0.0.1/", nil)
	r.Host = "127.0.0.1:8080"
	_ = GetCookieID(w, r)
	res := w.Result()
	defer res.Body.Close()
	var c *http.Cookie
	for _, cookie := range res.Cookies() {
		if cookie.Name == "ID" {
			c = cookie
			break
		}
	}
	if c == nil {
		t.Fatal("Set-Cookie header not found")
	}
	if c.Domain != "" {
		t.Errorf("Domain = %q, want empty", c.Domain)
	}
	if c.Secure {
		t.Error("Secure flag should be false on 127.0.0.1")
	}
}
