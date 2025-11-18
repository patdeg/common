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

// Package auth implements a minimal Google OAuth workflow used by the
// other packages in this repository. Users are redirected to Google for
// authentication and, on success, returned to /goog_callback where their
// email address is stored in a secure cookie. The optional "redirect"
// parameter is preserved as the OAuth state value so that users can be
// returned to the page they originally attempted to visit.
//
// To avoid open redirect vulnerabilities the state value is validated to
// ensure it is a relative path on the current site.
package auth

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/patdeg/common"
	"github.com/patdeg/common/gcp"
	"github.com/patdeg/common/track"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/appengine/v2/urlfetch"
)

var googleConfig = &oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
	Scopes: []string{
		"https://www.googleapis.com/auth/userinfo.email",
	},
	Endpoint: google.Endpoint,
}

var adminEmails map[string]bool

func init() {
	adminEmails = make(map[string]bool)
	for _, e := range strings.Split(os.Getenv("ADMIN_EMAILS"), ",") {
		e = strings.TrimSpace(e)
		if e != "" {
			adminEmails[e] = true
		}
	}
}

// sanitizeRedirect ensures the given value is a relative URL path. Any
// absolute URL or path not starting with "/" results in the fallback "/".
// The returned string may include query parameters.
func sanitizeRedirect(raw string) string {
	if raw == "" {
		return "/"
	}
	// Disallow any ".." segment after URL-decode
	u, err := url.PathUnescape(raw)
	if err != nil {
		return "/"
	}
	if strings.Contains(u, "..") {
		return "/"
	}
	// Must be relative and start with '/'
	if strings.HasPrefix(u, "/") && !strings.HasPrefix(u, "//") {
		return u
	}
	return "/"
}

// googleOAuthConfig returns an oauth2.Config with the RedirectURL set for the
// current request host. The redirect points to /goog_callback where the
// authorization code is exchanged.
func googleOAuthConfig(r *http.Request) *oauth2.Config {
	conf := *googleConfig
	conf.RedirectURL = "https://" + r.Host + "/goog_callback"
	return &conf
}

// RedirectIfNotLoggedIn checks for the login cookie and, if missing, initiates
// the OAuth login flow. The current path is used as the state value so users
// return to their original destination after logging in.
func RedirectIfNotLoggedIn(w http.ResponseWriter, r *http.Request) bool {
	if _, err := r.Cookie("user_email"); err != nil {
		url := googleOAuthConfig(r).AuthCodeURL(sanitizeRedirect(r.URL.Path))
		http.Redirect(w, r, url, http.StatusFound)
		return true
	}
	return false
}

// RedirectIfNotLoggedInAPI behaves like RedirectIfNotLoggedIn but returns a
// 401 status instead of redirecting. It is intended for API endpoints that
// require authentication.
func RedirectIfNotLoggedInAPI(w http.ResponseWriter, r *http.Request) bool {
	if _, err := r.Cookie("user_email"); err != nil {
		http.Error(w, "Login required", http.StatusUnauthorized)
		return true
	}
	return false
}

// GoogleLoginHandler starts the OAuth login process. The optional "redirect"
// parameter is validated and stored as the OAuth state to protect against open
// redirects. A login event is also recorded for analytics.
func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	state := sanitizeRedirect(r.FormValue("redirect"))
	url := googleOAuthConfig(r).AuthCodeURL(state)
	track.TrackEventDetails(w, r, common.GetCookieID(w, r), "Google Login", state, "", 0)
	http.Redirect(w, r, url, http.StatusFound)
}

// GoogleCallbackHandler handles the OAuth callback from Google. It exchanges
// the authorization code for a token, records the user email in a secure
// cookie and redirects the user back to the sanitized state value. If the
// state is empty or invalid a default dashboard or admin page is used.
func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := r.FormValue("code")
	state := sanitizeRedirect(r.FormValue("state"))
	conf := googleOAuthConfig(r)
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		common.Error("oauth exchange: %v", err)
		http.Error(w, "OAuth error", http.StatusInternalServerError)
		return
	}

	email, err := googleUserEmail(ctx, tok.AccessToken)
	if err != nil || email == "" {
		common.Error("googleUserEmail: %v", err)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	host := r.Host
	if h, _, err2 := net.SplitHostPort(host); err2 == nil {
		host = h
	}
	ck := &http.Cookie{
		Name:     "user_email",
		Value:    email,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	if host != "localhost" && host != "127.0.0.1" {
		ck.Domain = host
	}
	http.SetCookie(w, ck)

	role := "organizer"
	if adminEmails[email] {
		role = "admin"
	}
	if _, err := gcp.EnsureUserExists(ctx, email, role); err != nil {
		common.Error("EnsureUserExists: %v", err)
	}
	storedRole, err := gcp.GetUserRole(ctx, email)
	if err != nil {
		common.Error("GetUserRole: %v", err)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	redirect := state
	if redirect == "" || redirect == "/" {
		redirect = "/dashboard"
		if storedRole == "admin" {
			redirect = "/admin"
		}
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

// googleUserEmail fetches the authenticated user's email address using the
// provided OAuth access token. The token is passed via the Authorization header
// (not URL parameter) to prevent exposure in logs, proxies, and browser history.
// An empty string is returned if the request fails or the response cannot be decoded.
func googleUserEmail(c context.Context, token string) (string, error) {
	client := urlfetch.Client(c)

	// Create request with Authorization header instead of URL parameter
	// This prevents the access token from appearing in URLs, logs, or proxy caches
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		common.Error("Google userinfo request failed with status: %d", resp.StatusCode)
		return "", nil
	}

	var data struct{ Email string }
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	return data.Email, nil
}
