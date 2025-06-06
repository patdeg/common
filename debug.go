package common

import (
	"net/http"
	"net/http/httputil"

	"golang.org/x/net/context"
)

func DumpRequest(r *http.Request, withBody bool) {
	request, err := httputil.DumpRequest(r, withBody)
	if err != nil {
		Error("Error dumping request: %v", err)
		return
	}
	Debug("Request: %v", B2S(request))
}

func DumpRequestOut(r *http.Request, withBody bool) {
	request, err := httputil.DumpRequestOut(r, withBody)
	if err != nil {
		Error("Error dumping request: %v", err)
		return
	}
	Debug("Request: %v", B2S(request))
}

func DumpResponse(c context.Context, r *http.Response) {
	respDump, err := httputil.DumpResponse(r, true)
	if err != nil {
		Error("Error dumping response: %v", err)
		return
	}
	Debug("Response: %v", B2S(respDump))
}

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

func DumpCookies(r *http.Request) {
	for _, v := range r.Cookies() {
		Debug("Cookie %v = %v", v.Name, v.Value)
	}

}

func DebugInfo(r *http.Request) {
	// DebugInfo is disabled in non-App Engine environments

}
