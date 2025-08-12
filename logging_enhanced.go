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

// logging_enhanced.go provides enhanced logging with PII protection.
// This builds on top of the existing logging.go functions.

package common

import (
	"fmt"
	"os"
	"sync"

	"github.com/patdeg/common/logging"
)

var (
	// Global sanitizer instance
	globalSanitizer *logging.LogSanitizer
	sanitizerOnce   sync.Once

	// EnablePIIProtection controls whether PII sanitization is applied
	EnablePIIProtection = true
)

// initSanitizer initializes the global sanitizer
func initSanitizer() {
	sanitizerOnce.Do(func() {
		globalSanitizer = logging.NewLogSanitizer()

		// Configure based on environment
		if os.Getenv("LOG_PII_PROTECTION") == "false" {
			EnablePIIProtection = false
		}
	})
}

// DebugSafe writes a formatted debug message with PII sanitization.
// This is a PII-safe version of Debug that should be used when logging
// potentially sensitive information.
func DebugSafe(format string, v ...interface{}) {
	if !ISDEBUG {
		return
	}

	initSanitizer()

	message := fmt.Sprintf(format, v...)
	if EnablePIIProtection {
		message = globalSanitizer.Sanitize(message)
	}

	// Use the original Debug function with sanitized message
	Debug("%s", message)
}

// InfoSafe writes a formatted informational message with PII sanitization.
// This should be used instead of Info when the message might contain PII.
func InfoSafe(format string, v ...interface{}) {
	initSanitizer()

	message := fmt.Sprintf(format, v...)
	if EnablePIIProtection {
		message = globalSanitizer.Sanitize(message)
	}

	// Use the original Info function with sanitized message
	Info("%s", message)
}

// WarnSafe writes a formatted warning message with PII sanitization.
// This should be used instead of Warn when the message might contain PII.
func WarnSafe(format string, v ...interface{}) {
	initSanitizer()

	message := fmt.Sprintf(format, v...)
	if EnablePIIProtection {
		message = globalSanitizer.Sanitize(message)
	}

	// Use the original Warn function with sanitized message
	Warn("%s", message)
}

// ErrorSafe writes a formatted error message with PII sanitization.
// This should be used instead of Error when the message might contain PII.
func ErrorSafe(format string, v ...interface{}) {
	initSanitizer()

	message := fmt.Sprintf(format, v...)
	if EnablePIIProtection {
		message = globalSanitizer.Sanitize(message)
	}

	// Use the original Error function with sanitized message
	Error("%s", message)
}

// SanitizeMessage applies PII sanitization to a message.
// This can be used to sanitize messages before logging them.
func SanitizeMessage(message string) string {
	initSanitizer()

	if EnablePIIProtection {
		return globalSanitizer.Sanitize(message)
	}
	return message
}

// AddCustomPIIPattern adds a custom pattern for PII detection and sanitization.
// This allows applications to define their own PII patterns.
func AddCustomPIIPattern(name string, pattern string) error {
	initSanitizer()
	return globalSanitizer.AddCustomPattern(name, pattern)
}

// SetPIIProtection enables or disables PII protection globally
func SetPIIProtection(enabled bool) {
	EnablePIIProtection = enabled
}
