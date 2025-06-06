package common

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetCookieIDAttributes(t *testing.T) {
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
	if c.Domain != "example.com" {
		t.Errorf("Domain = %q, want example.com", c.Domain)
	}
}
