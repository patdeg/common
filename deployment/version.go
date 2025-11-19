package deployment

// Package deployment provides utilities for detecting deployment information
// in cloud environments (App Engine, Cloud Run, etc.).
// The version detection centralizes logic for cache-busting identifiers,
// logging, and error tracking across all AppEngine/Cloud Run projects.

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

var (
	versionOnce sync.Once
	versionID   string
)

// Get returns the detected deployment version. The detection runs only once per
// process to avoid repeated environment lookups.
//
// It checks multiple environment variables in order:
//   - DEMETERICS_VERSION (custom)
//   - DEPLOYMENT_VERSION (custom)
//   - VERSION_NAME (custom)
//   - GAE_VERSION (App Engine)
//   - GAE_DEPLOYMENT_ID (App Engine)
//   - K_REVISION (Cloud Run)
//   - K_SERVICE (Cloud Run)
//   - APP_VERSION (custom)
//
// If none are set, it falls back to "{ENVIRONMENT}-local" or "development-local".
func Get() string {
	versionOnce.Do(func() {
		versionID = detectVersion()
	})
	return versionID
}

func detectVersion() string {
	envKeys := []string{
		"DEMETERICS_VERSION",
		"DEPLOYMENT_VERSION",
		"VERSION_NAME",
		"GAE_VERSION",
		"GAE_DEPLOYMENT_ID",
		"K_REVISION",
		"K_SERVICE",
		"APP_VERSION",
	}

	for _, key := range envKeys {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}

	env := strings.TrimSpace(os.Getenv("ENVIRONMENT"))
	if env == "" {
		env = "development"
	}
	return fmt.Sprintf("%s-local", env)
}
