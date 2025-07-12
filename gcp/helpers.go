package gcp

import (
	"bytes"
	"log"
)

// VERSION stores the deployed application version retrieved from App Engine.
// It is set by the Version helper in appengine.go.
var VERSION string

// Debug writes a formatted debug message. This is a lightweight replacement
// for the helpers in the parent package to avoid an import cycle.
func Debug(format string, v ...interface{}) {
	log.Printf(format+"\n", v...)
}

// Info writes a formatted informational message.
func Info(format string, v ...interface{}) {
	log.Printf(format+"\n", v...)
}

// Error writes a formatted error message prefixed with "ERROR:".
func Error(format string, v ...interface{}) {
	log.Printf("ERROR: "+format+"\n", v...)
}

// b2s converts a byte slice possibly containing a null terminator into a string.
// It mirrors common.B2S but is duplicated here to avoid a package dependency.
func B2S(b []byte) string {
	n := bytes.Index(b, []byte{0})
	if n > 0 {
		return string(b[:n])
	}
	return string(b)
}
