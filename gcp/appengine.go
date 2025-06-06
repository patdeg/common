package gcp

import (
	"strings"

	"golang.org/x/net/context"
	appengine "google.golang.org/appengine/v2"
)

// Version retrieves the App Engine version ID from context and stores the
// major component in the exported `common.VERSION` variable. This helper should
// be called once during initialization so callers can read `common.VERSION`
// throughout the application.
func Version(c context.Context) string {
	version := appengine.VersionID(c)
	array := strings.Split(version, ".")
	VERSION = array[0]
	return VERSION
}
