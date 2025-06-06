package auth

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"mygotome/common"
	"mygotome/track"

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

func googleOAuthConfig(r *http.Request) *oauth2.Config {
	conf := *googleConfig
	conf.RedirectURL = "https://" + r.Host + "/goog_callback"
	return &conf
}

func RedirectIfNotLoggedIn(w http.ResponseWriter, r *http.Request) bool {
	if _, err := r.Cookie("organizer_email"); err != nil {
		url := googleOAuthConfig(r).AuthCodeURL(r.URL.Path)
		http.Redirect(w, r, url, http.StatusFound)
		return true
	}
	return false
}

func RedirectIfNotLoggedInAPI(w http.ResponseWriter, r *http.Request) bool {
	if _, err := r.Cookie("organizer_email"); err != nil {
		http.Error(w, "Login required", http.StatusUnauthorized)
		return true
	}
	return false
}

func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("redirect")
	if state == "" {
		state = "/"
	}
	url := googleOAuthConfig(r).AuthCodeURL(state)
	track.TrackEventDetails(w, r, common.GetCookieID(w, r), "Google Login", state, "", 0)
	http.Redirect(w, r, url, http.StatusFound)
}

func GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	code := r.FormValue("code")
	state := r.FormValue("state")
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
		Name:     "organizer_email",
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
	if _, err := common.EnsureUserExists(ctx, email, role); err != nil {
		common.Error("EnsureUserExists: %v", err)
	}
	storedRole, err := common.GetUserRole(ctx, email)
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

func googleUserEmail(c context.Context, token string) (string, error) {
	client := urlfetch.Client(c)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var data struct{ Email string }
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	return data.Email, nil
}
