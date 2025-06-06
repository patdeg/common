// Package main provides a minimal example server showing how to use the
// authentication helpers from the common repository. It defines a simple
// greeting handler and wires up the OAuth login and callback handlers from the
// auth subpackage.
//
// Environment variables:
//   - GOOGLE_OAUTH_CLIENT_ID and GOOGLE_OAUTH_CLIENT_SECRET must be set with
//     credentials for a Google OAuth web client.
//   - ADMIN_EMAILS optionally provides a comma separated list of admin
//     accounts allowed to log in.
//   - PORT specifies the HTTP port for the server (default 8080).
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/patdeg/common"
	"github.com/patdeg/common/auth"
)

// HelloHandler responds with a greeting to authenticated users. If the user is
// not logged in, they are redirected to the Google OAuth login flow.
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	if auth.RedirectIfNotLoggedIn(w, r) {
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintln(w, "<h1>Hello World</h1>")
}

func main() {
	// The root path requires login and displays a greeting.
	http.HandleFunc("/", HelloHandler)

	// GoogleLoginHandler initiates the OAuth flow by redirecting the user
	// to Google's consent page.
	http.HandleFunc("/goog_login", auth.GoogleLoginHandler)

	// GoogleCallbackHandler completes the OAuth flow and sets the login
	// cookie before redirecting back to the requested page.
	http.HandleFunc("/goog_callback", auth.GoogleCallbackHandler)

	// The HTTP server listens on the port specified by the PORT environment
	// variable. It defaults to 8080 when unset.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	common.Info("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
