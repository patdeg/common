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

package auth

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestSanitizeRedirect(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", "/"},
		{"/dashboard", "/dashboard"},
		{"http://evil.com", "/"},
		{"//evil", "/"},
		{"/path?x=1", "/path?x=1"},
	}
	for _, c := range cases {
		if got := sanitizeRedirect(c.in); got != c.want {
			t.Errorf("sanitizeRedirect(%q) = %q; want %q", c.in, got, c.want)
		}
	}
}

func TestGoogleLoginHandlerUnsafeRedirect(t *testing.T) {
	r := httptest.NewRequest("GET", "https://example.com/login?redirect=http://evil.com", nil)
	r.Host = "example.com"
	w := httptest.NewRecorder()
	GoogleLoginHandler(w, r)
	res := w.Result()
	loc := res.Header.Get("Location")
	if loc == "" {
		t.Fatal("no redirect")
	}
	u, err := url.Parse(loc)
	if err != nil {
		t.Fatalf("invalid redirect URL: %v", err)
	}
	if !strings.Contains(u.RawQuery, "state=%2F") {
		t.Errorf("state not sanitized in redirect: %s", loc)
	}
}
