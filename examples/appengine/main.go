package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/patdeg/common"
	"github.com/patdeg/common/auth"
)

// HelloHandler greets authenticated users.
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	if auth.RedirectIfNotLoggedIn(w, r) {
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintln(w, "<h1>Hello World</h1>")
}

func main() {
	http.HandleFunc("/", HelloHandler)
	http.HandleFunc("/goog_login", auth.GoogleLoginHandler)
	http.HandleFunc("/goog_callback", auth.GoogleCallbackHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	common.Info("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
