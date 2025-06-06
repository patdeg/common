package common

import "testing"

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
