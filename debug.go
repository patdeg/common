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
		Info("Cookie:")
		Info("  - Name: %v", cookie.Name)
		Info("  - Value: %v", cookie.Value)
		Info("  - Path: %v", cookie.Path)
		Info("  - Domain: %v", cookie.Domain)
		Info("  - Expires: %v", cookie.Expires)
		Info("  - RawExpires: %v", cookie.RawExpires)
		Info("  - MaxAge: %v", cookie.MaxAge)
		Info("  - Secure:%v", cookie.Secure)
		Info("  - HttpOnly: %v", cookie.HttpOnly)
		Info("  - Raw: %v", cookie.Raw)
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
