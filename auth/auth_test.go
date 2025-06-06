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
