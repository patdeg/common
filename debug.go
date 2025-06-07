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

// Package common provides assorted helpers. This file contains debugging
// utilities used to dump HTTP requests, responses and cookies when
// troubleshooting issues during development.
package common

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"

	"golang.org/x/net/context"
)

// DumpRequest logs the incoming HTTP request at debug level. If withBody is
// true, the request body is included and closed after logging. This helper
// should only be used during development as it may expose sensitive data.
func DumpRequest(r *http.Request, withBody bool) {
	request, err := httputil.DumpRequest(r, withBody)
	if err != nil {
		Error("Error dumping request: %v", err)
		return
	}
	Debug("Request: %v", B2S(request))
	if withBody && r.Body != nil {
		r.Body.Close()
	}
}

// DumpRequestOut logs an outbound client request. If withBody is true the
// request body is included. Only use in non-production environments as
// headers or bodies may contain private information.
func DumpRequestOut(r *http.Request, withBody bool) {
	request, err := httputil.DumpRequestOut(r, withBody)
	if err != nil {
		Error("Error dumping request: %v", err)
		return
	}
	Debug("Request: %v", B2S(request))
}

// DumpResponse logs an HTTP response. The response body is left open so callers
// may still read it and must be closed by the caller. Avoid using this helper in
// production if the response contains sensitive data.
func DumpResponse(c context.Context, r *http.Response) {
	if r == nil {
		return
	}
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		Error("Error dumping response: %v", err)
		return
	}
	r.Body.Close()

	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	respDump, err := httputil.DumpResponse(r, true)
	if err != nil {
		Error("Error dumping response: %v", err)
		return
	}
	Debug("Response: %v", B2S(respDump))

	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
}

// DumpCookie logs the details of a single cookie for debugging. Do not log
// cookies in production as they may contain session or tracking information.
func DumpCookie(c context.Context, cookie *http.Cookie) {
	if cookie != nil {
		Debug("Cookie:")
		Debug("  - Name: %v", cookie.Name)
		Debug("  - Value: %v", cookie.Value)
		Debug("  - Path: %v", cookie.Path)
		Debug("  - Domain: %v", cookie.Domain)
		Debug("  - Expires: %v", cookie.Expires)
		Debug("  - RawExpires: %v", cookie.RawExpires)
		Debug("  - MaxAge: %v", cookie.MaxAge)
		Debug("  - Secure:%v", cookie.Secure)
		Debug("  - HttpOnly: %v", cookie.HttpOnly)
		Debug("  - Raw: %v", cookie.Raw)
	} else {
		Debug("Cookie is null")
	}
}

// DumpCookies logs all cookies present on the request. Use with caution and
// never in production if cookies contain private data.
func DumpCookies(r *http.Request) {
	for _, v := range r.Cookies() {
		Debug("Cookie %v = %v", v.Name, v.Value)
	}

}

func DebugInfo(r *http.Request) {
	// DebugInfo is disabled in non-App Engine environments

}
