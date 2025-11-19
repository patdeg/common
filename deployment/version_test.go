package deployment

import (
	"os"
	"strings"
	"sync"
	"testing"
)

func TestGet(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected string
	}{
		{
			name: "DEMETERICS_VERSION takes priority",
			envVars: map[string]string{
				"DEMETERICS_VERSION": "custom-v1.2.3",
				"GAE_VERSION":        "gae-v1",
				"K_REVISION":         "cloudrun-rev1",
			},
			expected: "custom-v1.2.3",
		},
		{
			name: "DEPLOYMENT_VERSION used when DEMETERICS_VERSION not set",
			envVars: map[string]string{
				"DEPLOYMENT_VERSION": "deploy-v2.0.0",
				"GAE_VERSION":        "gae-v1",
			},
			expected: "deploy-v2.0.0",
		},
		{
			name: "VERSION_NAME used when higher priority vars not set",
			envVars: map[string]string{
				"VERSION_NAME": "v3.0.0",
				"GAE_VERSION":  "gae-v1",
			},
			expected: "v3.0.0",
		},
		{
			name: "GAE_VERSION used for App Engine",
			envVars: map[string]string{
				"GAE_VERSION": "prod-55",
			},
			expected: "prod-55",
		},
		{
			name: "GAE_DEPLOYMENT_ID used for App Engine",
			envVars: map[string]string{
				"GAE_DEPLOYMENT_ID": "deployment-abc123",
			},
			expected: "deployment-abc123",
		},
		{
			name: "K_REVISION used for Cloud Run",
			envVars: map[string]string{
				"K_REVISION": "myservice-00042-abc",
			},
			expected: "myservice-00042-abc",
		},
		{
			name: "K_SERVICE used for Cloud Run",
			envVars: map[string]string{
				"K_SERVICE": "my-cloud-run-service",
			},
			expected: "my-cloud-run-service",
		},
		{
			name: "APP_VERSION used as fallback",
			envVars: map[string]string{
				"APP_VERSION": "app-v1.0.0",
			},
			expected: "app-v1.0.0",
		},
		{
			name: "Falls back to ENVIRONMENT-local when no version vars set",
			envVars: map[string]string{
				"ENVIRONMENT": "production",
			},
			expected: "production-local",
		},
		{
			name: "Falls back to development-local when no vars set",
			envVars: map[string]string{
				// No environment variables set
			},
			expected: "development-local",
		},
		{
			name: "Trims whitespace from environment variables",
			envVars: map[string]string{
				"GAE_VERSION": "  prod-55  ",
			},
			expected: "prod-55",
		},
		{
			name: "Empty string in higher priority skips to next",
			envVars: map[string]string{
				"DEMETERICS_VERSION": "",
				"DEPLOYMENT_VERSION": "",
				"GAE_VERSION":        "prod-55",
			},
			expected: "prod-55",
		},
		{
			name: "Whitespace-only string in higher priority skips to next",
			envVars: map[string]string{
				"DEMETERICS_VERSION": "   ",
				"GAE_VERSION":        "prod-55",
			},
			expected: "prod-55",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all relevant environment variables
			envKeys := []string{
				"DEMETERICS_VERSION",
				"DEPLOYMENT_VERSION",
				"VERSION_NAME",
				"GAE_VERSION",
				"GAE_DEPLOYMENT_ID",
				"K_REVISION",
				"K_SERVICE",
				"APP_VERSION",
				"ENVIRONMENT",
			}
			for _, key := range envKeys {
				os.Unsetenv(key)
			}

			// Set test-specific environment variables
			for key, val := range tt.envVars {
				os.Setenv(key, val)
			}

			// Reset the sync.Once so we can test detection again
			versionOnce = *new(sync.Once)
			versionID = ""

			got := Get()
			if got != tt.expected {
				t.Errorf("Get() = %q, want %q", got, tt.expected)
			}

			// Verify idempotency - calling again should return same value
			got2 := Get()
			if got2 != got {
				t.Errorf("Get() not idempotent: first call = %q, second call = %q", got, got2)
			}

			// Clean up
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}

func TestDetectVersion(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected string
	}{
		{
			name: "Priority order test",
			envVars: map[string]string{
				"DEMETERICS_VERSION": "priority1",
				"DEPLOYMENT_VERSION": "priority2",
				"VERSION_NAME":       "priority3",
				"GAE_VERSION":        "priority4",
				"GAE_DEPLOYMENT_ID":  "priority5",
				"K_REVISION":         "priority6",
				"K_SERVICE":          "priority7",
				"APP_VERSION":        "priority8",
			},
			expected: "priority1",
		},
		{
			name: "Skip to second priority",
			envVars: map[string]string{
				"DEPLOYMENT_VERSION": "priority2",
				"VERSION_NAME":       "priority3",
			},
			expected: "priority2",
		},
		{
			name: "Staging environment fallback",
			envVars: map[string]string{
				"ENVIRONMENT": "staging",
			},
			expected: "staging-local",
		},
		{
			name: "Dev environment fallback",
			envVars: map[string]string{
				"ENVIRONMENT": "dev",
			},
			expected: "dev-local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all relevant environment variables
			envKeys := []string{
				"DEMETERICS_VERSION",
				"DEPLOYMENT_VERSION",
				"VERSION_NAME",
				"GAE_VERSION",
				"GAE_DEPLOYMENT_ID",
				"K_REVISION",
				"K_SERVICE",
				"APP_VERSION",
				"ENVIRONMENT",
			}
			for _, key := range envKeys {
				os.Unsetenv(key)
			}

			// Set test-specific environment variables
			for key, val := range tt.envVars {
				os.Setenv(key, val)
			}

			got := detectVersion()
			if got != tt.expected {
				t.Errorf("detectVersion() = %q, want %q", got, tt.expected)
			}

			// Clean up
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}

func TestGetCacheBusting(t *testing.T) {
	// Set up a known version
	os.Setenv("GAE_VERSION", "prod-42")
	defer os.Unsetenv("GAE_VERSION")

	// Reset the sync.Once
	versionOnce = *new(sync.Once)
	versionID = ""

	version := Get()

	// Verify it's suitable for cache busting
	if version == "" {
		t.Error("Get() returned empty string, not suitable for cache busting")
	}

	if strings.Contains(version, " ") {
		t.Error("Get() returned version with spaces, not suitable for URLs")
	}
}
