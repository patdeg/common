package common

import (
	"strings"

	"golang.org/x/net/context"
	appengine "google.golang.org/appengine/v2"
)

func Version(c context.Context) string {
	version := appengine.VersionID(c)
	array := strings.Split(version, ".")
	VERSION = array[0]
	return VERSION
}
