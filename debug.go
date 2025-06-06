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

func DumpCookies(r *http.Request) {
	for _, v := range r.Cookies() {
		Debug("Cookie %v = %v", v.Name, v.Value)
	}

}

func DebugInfo(r *http.Request) {
	// DebugInfo is disabled in non-App Engine environments

}
