package common

import "net/url"

// IsValidHTTPURL verifies that dest is an absolute HTTP or HTTPS URL.
func IsValidHTTPURL(dest string) bool {
	if dest == "" {
		return false
	}
	u, err := url.Parse(dest)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	if u.Host == "" {
		return false
	}
	return true
}
