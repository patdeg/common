package common

import "net/url"

// Utilities for validating URLs used by callers. The goal is to guard against
// misconfigured links and open redirects by ensuring only absolute HTTP or HTTPS
// destinations are accepted. Additional checks are performed to verify the
// scheme and host are well-formed before using the URL.

// IsValidHTTPURL verifies that dest is an absolute HTTP or HTTPS URL.
func IsValidHTTPURL(dest string) bool {
	// Empty strings are never valid URLs.
	if dest == "" {
		return false
	}
	// url.Parse handles relative paths and malformed URLs. If parsing fails we
	// know the input is not usable as a link.
	u, err := url.Parse(dest)
	if err != nil {
		return false
	}
	// Only allow the http or https schemes to avoid unsupported or unsafe
	// protocols.
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	// A missing host indicates a relative URL or an incorrectly formatted URL
	// such as "http://" with no hostname.
	if u.Host == "" {
		return false
	}
	return true
}
