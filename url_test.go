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

// Package common contains tests for URL helpers.
package common

import "testing"

// TestIsValidHTTPURL checks that IsValidHTTPURL accepts HTTP/HTTPS URLs and
// rejects other schemes or malformed strings.

func TestIsValidHTTPURL(t *testing.T) {
	valid := []string{
		"http://example.com",
		"https://example.com/path",
		"http://example.com:8080",
		"https://example.com?q=1",
	}
	for _, u := range valid {
		if !IsValidHTTPURL(u) {
			t.Errorf("expected valid URL: %s", u)
		}
	}
	invalid := []string{
		"",
		"ftp://example.com",
		"/relative",
		"http://",
		"https://",
		"http:/example.com",
		"http:example.com",
		"file:///etc/passwd",
	}
	for _, u := range invalid {
		if IsValidHTTPURL(u) {
			t.Errorf("expected invalid URL: %s", u)
		}
	}
}

// TestNormalizeBase checks that NormalizeBase correctly adds https:// prefix
// and removes trailing slashes.
func TestNormalizeBase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"  ", ""},
		{"example.com", "https://example.com"},
		{"example.com/", "https://example.com"},
		{"http://example.com", "http://example.com"},
		{"http://example.com/", "http://example.com"},
		{"https://example.com", "https://example.com"},
		{"https://example.com/", "https://example.com"},
		{"https://example.com/path", "https://example.com/path"},
		{"https://example.com/path/", "https://example.com/path"},
		{"  example.com  ", "https://example.com"},
		{"example.com:8080", "https://example.com:8080"},
		{"http://example.com:8080/", "http://example.com:8080"},
	}
	for _, tt := range tests {
		got := NormalizeBase(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeBase(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// TestJoin checks that Join correctly concatenates base URLs and paths.
func TestJoin(t *testing.T) {
	tests := []struct {
		base string
		path string
		want string
	}{
		{"", "", ""},
		{"", "/path", ""},
		{"example.com", "", "https://example.com"},
		{"example.com", "/path", "https://example.com/path"},
		{"example.com", "path", "https://example.com/path"},
		{"example.com/", "/path", "https://example.com/path"},
		{"https://example.com", "/path", "https://example.com/path"},
		{"https://example.com/", "/path", "https://example.com/path"},
		{"https://example.com", "path", "https://example.com/path"},
		{"https://example.com/base", "/path", "https://example.com/base/path"},
		{"https://example.com/base/", "/path", "https://example.com/base/path"},
		{"http://example.com:8080", "/api/v1", "http://example.com:8080/api/v1"},
		{"http://example.com:8080/", "/api/v1", "http://example.com:8080/api/v1"},
		{"https://example.com/api", "/users", "https://example.com/api/users"},
		{"https://example.com/api/", "/users", "https://example.com/api/users"},
		{"  example.com  ", "  /path  ", "https://example.com/path"},
	}
	for _, tt := range tests {
		got := Join(tt.base, tt.path)
		if got != tt.want {
			t.Errorf("Join(%q, %q) = %q, want %q", tt.base, tt.path, got, tt.want)
		}
	}
}
