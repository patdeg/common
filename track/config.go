// Package track contains analytics helpers for recording visits and events.
package track

import "os"

// getEnv returns the value of the environment variable or the provided
// default when the variable is unset. For example:
//
//	port := getEnv("PORT", "8080")
//
// will read the PORT environment variable and fall back to "8080" when it is
// not present.
func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
