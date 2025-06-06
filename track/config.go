// Package track contains analytics helpers for recording visits and events.
package track

import "os"

// getEnv returns the value of the environment variable or a default.
func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
