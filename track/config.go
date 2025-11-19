// Copyright 2025 Patrick Deglon
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

// ConfigureTouchpoints overrides the default BigQuery project and dataset used
// for touch point events. Empty values are ignored, allowing callers to update
// only one of the parameters.
//
// Typical usage at application startup:
//
//	track.ConfigureTouchpoints("my-project", "marketing_touchpoints")
func ConfigureTouchpoints(projectID, datasetID string) {
	if projectID != "" {
		touchpointsProjectID = projectID
	}
	if datasetID != "" {
		touchpointsDataset = datasetID
	}
}
