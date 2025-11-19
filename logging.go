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

// logging.go provides a tiny wrapper around the standard log package.
// The helpers here format messages consistently and funnel all logs
// through log.Printf. Debug obeys the global ISDEBUG variable so it
// can be disabled globally.

package common

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/datastore"
)

var (
	// ERROR_DATASTORE_ENTITY holds the Datastore entity name for error logging.
	// When set, the Error() function will store errors in Datastore.
	// Example: "Error" for production, "UAT_Error" for UAT, "DEV_Error" for development
	ERROR_DATASTORE_ENTITY string

	// Global datastore client for error logging
	errorClient *datastore.Client
)

// ErrorEntity represents an error entry stored in Datastore
type ErrorEntity struct {
	Timestamp      time.Time `datastore:"timestamp"`
	Message        string    `datastore:"message"`
	GAEApplication string    `datastore:"gae_application,omitempty"`
	GAEService     string    `datastore:"gae_service,omitempty"`
	GAEVersion     string    `datastore:"gae_version,omitempty"`
	GAEInstance    string    `datastore:"gae_instance,omitempty"`
	GAEMemoryMB    string    `datastore:"gae_memory_mb,omitempty"`
	GAEEnv         string    `datastore:"gae_env,omitempty"`
	GAERuntime     string    `datastore:"gae_runtime,omitempty"`
	ProjectID      string    `datastore:"project_id,omitempty"`
	Environment    string    `datastore:"environment,omitempty"`
}

// getAppEngineMetadata collects useful App Engine runtime environment variables
func getAppEngineMetadata() ErrorEntity {
	return ErrorEntity{
		GAEApplication: os.Getenv("GAE_APPLICATION"),
		GAEService:     os.Getenv("GAE_SERVICE"),
		GAEVersion:     os.Getenv("GAE_VERSION"),
		GAEInstance:    os.Getenv("GAE_INSTANCE"),
		GAEMemoryMB:    os.Getenv("GAE_MEMORY_MB"),
		GAEEnv:         os.Getenv("GAE_ENV"),
		GAERuntime:     os.Getenv("GAE_RUNTIME"),
		ProjectID:      os.Getenv("PROJECT_ID"),
		Environment:    os.Getenv("APP_ENV"),
	}
}

// InitErrorDatastore initializes the error logging datastore client
func InitErrorDatastore() error {
	if ERROR_DATASTORE_ENTITY == "" {
		return nil // No entity name set, skip initialization
	}

	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}
	if projectID == "" {
		return fmt.Errorf("PROJECT_ID not configured")
	}

	ctx := context.Background()
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to create datastore client: %v", err)
	}
	errorClient = client
	return nil
}

// Debug writes a formatted debug message when ISDEBUG is true.
// A newline is appended so callers do not have to include one.
func Debug(format string, v ...interface{}) {
	if !ISDEBUG {
		return
	}
	// Include trailing newline for consistency with other helpers.
	log.Printf(format+"\n", v...)
}

// Info writes a formatted informational message.
// A newline is appended to keep log lines consistent.
func Info(format string, v ...interface{}) {
	log.Printf(format+"\n", v...)
}

// Warn writes a formatted warning message with an "WARNING:" prefix.
// The prefix helps grep for warnings in log files.
func Warn(format string, v ...interface{}) {
	log.Printf("WARNING: "+format+"\n", v...)
}

// Error writes a formatted error message with an "ERROR:" prefix.
// The prefix helps grep for errors in log files.
// If ERROR_DATASTORE_ENTITY is set, also stores the error in Datastore.
func Error(format string, v ...interface{}) {
	errorMsg := fmt.Sprintf(format, v...)
	log.Printf("ERROR: %s\n", errorMsg)

	// Store in Datastore if configured
	if ERROR_DATASTORE_ENTITY != "" && errorClient != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Create error entity with metadata
			errorEntry := getAppEngineMetadata()
			errorEntry.Timestamp = time.Now()
			errorEntry.Message = errorMsg

			// Use timestamp as key for uniqueness
			keyName := fmt.Sprintf("%d", time.Now().UnixNano())
			key := datastore.NameKey(ERROR_DATASTORE_ENTITY, keyName, nil)

			// Store in Datastore (non-blocking)
			if _, err := errorClient.Put(ctx, key, &errorEntry); err != nil {
				// Log to stdout if Datastore storage fails, but don't recurse
				log.Printf("WARNING: Failed to store error in Datastore: %v\n", err)
			}
		}()
	}
}

// Fatal logs an error message with "FATAL: " prefix and exits the program.
// This is used for unrecoverable errors during startup or critical failures.
// The function logs the message and then calls os.Exit(1).
func Fatal(format string, v ...interface{}) {
	errorMsg := fmt.Sprintf(format, v...)
	log.Printf("FATAL: %s\n", errorMsg)

	// Store in Datastore if configured (best effort, don't wait)
	if ERROR_DATASTORE_ENTITY != "" && errorClient != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			// Create error entity with metadata
			errorEntry := getAppEngineMetadata()
			errorEntry.Timestamp = time.Now()
			errorEntry.Message = "FATAL: " + errorMsg

			// Use timestamp as key for uniqueness
			keyName := fmt.Sprintf("%d", time.Now().UnixNano())
			key := datastore.NameKey(ERROR_DATASTORE_ENTITY, keyName, nil)

			// Store in Datastore (fire and forget)
			if _, err := errorClient.Put(ctx, key, &errorEntry); err != nil {
				log.Printf("WARNING: Failed to store error in Datastore: %v\n", err)
			}
		}()
	}

	// Give a brief moment for the log to be written
	time.Sleep(100 * time.Millisecond)
	os.Exit(1)
}
