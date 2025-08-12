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

package logging

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

// LogSanitizer sanitizes log messages to remove PII
type LogSanitizer struct {
	mu       sync.RWMutex
	patterns map[string]*regexp.Regexp

	// Configuration options
	maskEmails     bool
	maskIPs        bool
	maskCreditCard bool
	maskSSN        bool
	maskPhone      bool
	maskCustom     map[string]*regexp.Regexp
}

// NewLogSanitizer creates a new log sanitizer with default settings
func NewLogSanitizer() *LogSanitizer {
	ls := &LogSanitizer{
		patterns:       make(map[string]*regexp.Regexp),
		maskEmails:     true,
		maskIPs:        true,
		maskCreditCard: true,
		maskSSN:        true,
		maskPhone:      true,
		maskCustom:     make(map[string]*regexp.Regexp),
	}

	// Initialize default patterns
	ls.initPatterns()

	return ls
}

// initPatterns initializes the default regex patterns for PII detection
func (ls *LogSanitizer) initPatterns() {
	// Email pattern
	ls.patterns["email"] = regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)

	// IP address patterns (both IPv4 and IPv6)
	ls.patterns["ipv4"] = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	ls.patterns["ipv6"] = regexp.MustCompile(`\b(?:[A-Fa-f0-9]{1,4}:){7}[A-Fa-f0-9]{1,4}\b`)

	// Credit card pattern (basic - matches 13-19 digit sequences)
	ls.patterns["creditcard"] = regexp.MustCompile(`\b(?:\d[ -]*?){13,19}\b`)

	// SSN pattern (US Social Security Number)
	ls.patterns["ssn"] = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b|\b\d{9}\b`)

	// Phone number patterns (various formats)
	ls.patterns["phone"] = regexp.MustCompile(`\b(?:\+?1[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}\b`)

	// API key patterns (common formats)
	ls.patterns["apikey"] = regexp.MustCompile(`\b(api[_-]?key|apikey|api_secret|api[_-]?token)[\s]*[:=][\s]*['"]?[\w-]{20,}['"]?\b`)

	// JWT token pattern
	ls.patterns["jwt"] = regexp.MustCompile(`\beyJ[A-Za-z0-9-_]+\.eyJ[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\b`)

	// Password in query strings or JSON
	ls.patterns["password"] = regexp.MustCompile(`(password|passwd|pwd|pass)[\s]*[:=][\s]*['"]?[^'"\s,}]+['"]?`)
}

// Sanitize removes or masks PII from a log message
func (ls *LogSanitizer) Sanitize(message string) string {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	// Apply email masking
	if ls.maskEmails {
		message = ls.patterns["email"].ReplaceAllStringFunc(message, maskEmail)
	}

	// Apply IP masking
	if ls.maskIPs {
		message = ls.patterns["ipv4"].ReplaceAllString(message, "xxx.xxx.xxx.xxx")
		message = ls.patterns["ipv6"].ReplaceAllString(message, "xxxx:xxxx:xxxx:xxxx:xxxx:xxxx:xxxx:xxxx")
	}

	// Apply credit card masking
	if ls.maskCreditCard {
		message = ls.patterns["creditcard"].ReplaceAllStringFunc(message, maskCreditCard)
	}

	// Apply SSN masking
	if ls.maskSSN {
		message = ls.patterns["ssn"].ReplaceAllString(message, "xxx-xx-xxxx")
	}

	// Apply phone masking
	if ls.maskPhone {
		message = ls.patterns["phone"].ReplaceAllStringFunc(message, maskPhone)
	}

	// Always mask API keys and passwords
	message = ls.patterns["apikey"].ReplaceAllString(message, "$1=***REDACTED***")
	message = ls.patterns["jwt"].ReplaceAllString(message, "***JWT_REDACTED***")
	message = ls.patterns["password"].ReplaceAllString(message, "$1=***REDACTED***")

	// Apply custom patterns
	for name, pattern := range ls.maskCustom {
		message = pattern.ReplaceAllString(message, fmt.Sprintf("***%s_REDACTED***", strings.ToUpper(name)))
	}

	return message
}

// SanitizeStruct sanitizes a struct by masking PII in string fields
func (ls *LogSanitizer) SanitizeStruct(data interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Use reflection to iterate through struct fields
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		// If not a struct, try to marshal and return as-is
		result["value"] = ls.sanitizeValue(data)
		return result
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Skip unexported fields
		if !value.CanInterface() {
			continue
		}

		fieldName := field.Name
		fieldValue := value.Interface()

		// Check if field name suggests PII
		lowerName := strings.ToLower(fieldName)
		if containsPIIFieldName(lowerName) {
			result[fieldName] = "***REDACTED***"
		} else {
			result[fieldName] = ls.sanitizeValue(fieldValue)
		}
	}

	return result
}

// sanitizeValue sanitizes a single value
func (ls *LogSanitizer) sanitizeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return ls.Sanitize(v)
	case []byte:
		return ls.Sanitize(string(v))
	case fmt.Stringer:
		return ls.Sanitize(v.String())
	default:
		// For complex types, marshal to JSON and sanitize
		if data, err := json.Marshal(v); err == nil {
			sanitized := ls.Sanitize(string(data))
			var result interface{}
			if json.Unmarshal([]byte(sanitized), &result) == nil {
				return result
			}
		}
		return value
	}
}

// AddCustomPattern adds a custom pattern for sanitization
func (ls *LogSanitizer) AddCustomPattern(name string, pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}

	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.maskCustom[name] = re

	return nil
}

// SetEmailMasking enables or disables email masking
func (ls *LogSanitizer) SetEmailMasking(enabled bool) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.maskEmails = enabled
}

// SetIPMasking enables or disables IP address masking
func (ls *LogSanitizer) SetIPMasking(enabled bool) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.maskIPs = enabled
}

// Helper functions for masking

// maskEmail partially masks an email address
func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***@***.***"
	}

	local := parts[0]
	domain := parts[1]

	// Mask local part
	if len(local) > 2 {
		local = local[:1] + "***" + local[len(local)-1:]
	} else {
		local = "***"
	}

	// Partially mask domain
	domainParts := strings.Split(domain, ".")
	if len(domainParts) > 1 {
		domainParts[0] = "***"
	}

	return local + "@" + strings.Join(domainParts, ".")
}

// maskCreditCard masks a credit card number, showing only last 4 digits
func maskCreditCard(cc string) string {
	// Remove spaces and dashes
	cleaned := strings.ReplaceAll(strings.ReplaceAll(cc, " ", ""), "-", "")
	if len(cleaned) > 4 {
		return "****-****-****-" + cleaned[len(cleaned)-4:]
	}
	return "****-****-****-****"
}

// maskPhone masks a phone number
func maskPhone(phone string) string {
	// Keep country code and area code structure if present
	if strings.HasPrefix(phone, "+") {
		return "+*-***-***-****"
	}
	return "***-***-****"
}

// containsPIIFieldName checks if a field name likely contains PII
func containsPIIFieldName(name string) bool {
	piiFields := []string{
		"email", "mail", "password", "passwd", "pwd", "pass",
		"ssn", "social", "phone", "mobile", "cell",
		"credit", "card", "cvv", "cvc",
		"api", "key", "token", "secret",
		"ip", "address", "addr",
		"name", "firstname", "lastname", "username",
		"dob", "birth", "age",
		"salary", "income", "wage",
	}

	for _, field := range piiFields {
		if strings.Contains(name, field) {
			return true
		}
	}

	return false
}
