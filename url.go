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

package common

import (
	"net/url"
	"strings"
)

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

// NormalizeBase returns a sanitized base URL with scheme and no trailing slash.
// If the input is empty, an empty string is returned.
// If the input lacks an http:// or https:// prefix, https:// is added.
func NormalizeBase(raw string) string {
	base := strings.TrimSpace(raw)
	if base == "" {
		return ""
	}
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "https://" + base
	}
	return strings.TrimSuffix(base, "/")
}

// Join concatenates the provided path with the normalized base URL.
// The base URL is normalized using NormalizeBase, and the path is joined safely.
// If the base is empty after normalization, an empty string is returned.
// If the path is empty, the normalized base is returned.
func Join(rawBase, path string) string {
	base := NormalizeBase(rawBase)
	if base == "" {
		return ""
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return base
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u, err := url.Parse(base)
	if err != nil {
		return base + path
	}
	if u.Path == "" || u.Path == "/" {
		u.Path = path
	} else {
		u.Path = strings.TrimRight(u.Path, "/") + path
	}
	return u.String()
}
