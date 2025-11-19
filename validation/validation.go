package validation

// Package validation provides reusable validation functions for API requests.
//
// This package centralizes common validation logic to ensure consistent error
// messages and reduce code duplication across API handlers.
//
// Design principles:
// - Each validator returns a ValidationError with a helpful message
// - Validators are composable (can be chained together)
// - Error messages are actionable and don't leak sensitive information
// - All validators are safe for concurrent use

import (
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

// ValidationError represents a validation failure with a helpful message.
type ValidationError struct {
	Field   string // The field that failed validation
	Message string // Human-readable error message
	Code    string // Machine-readable error code (e.g., "required", "invalid_format")
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}

// ValidationErrors represents multiple validation failures.
type ValidationErrors []ValidationError

// Error implements the error interface.
func (errs ValidationErrors) Error() string {
	if len(errs) == 0 {
		return ""
	}
	if len(errs) == 1 {
		return errs[0].Error()
	}
	var messages []string
	for _, err := range errs {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// HasErrors returns true if there are any validation errors.
func (errs ValidationErrors) HasErrors() bool {
	return len(errs) > 0
}

// Regular expressions for common validation patterns
var (
	// alphanumericDashUnderscore matches alphanumeric characters, dashes, and underscores
	alphanumericDashUnderscore = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	// alphanumericSpaceDashUnderscore matches alphanumeric characters, spaces, dashes, and underscores
	alphanumericSpaceDashUnderscore = regexp.MustCompile(`^[a-zA-Z0-9 _-]+$`)

	// uuidPattern matches UUID v4 format
	uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

	// modelNamePattern matches typical AI model names (e.g., "gpt-4o", "llama-3.1-8b")
	modelNamePattern = regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)

	// cronPattern matches basic cron expressions (not exhaustive)
	cronPattern = regexp.MustCompile(`^(\*|([0-9]|1[0-9]|2[0-9]|3[0-9]|4[0-9]|5[0-9])|\*\/([0-9]|1[0-9]|2[0-9]|3[0-9]|4[0-9]|5[0-9])) (\*|([0-9]|1[0-9]|2[0-3])|\*\/([0-9]|1[0-9]|2[0-3])) (\*|([1-9]|1[0-9]|2[0-9]|3[0-1])|\*\/([1-9]|1[0-9]|2[0-9]|3[0-1])) (\*|([1-9]|1[0-2])|\*\/([1-9]|1[0-2])) (\*|([0-6])|\*\/([0-6]))$`)
)

// Required validates that a string field is not empty.
func Required(field, value string) *ValidationError {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{
			Field:   field,
			Message: "is required",
			Code:    "required",
		}
	}
	return nil
}

// MaxLength validates that a string does not exceed the maximum length.
func MaxLength(field, value string, max int) *ValidationError {
	if len(value) > max {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must not exceed %d characters (got %d)", max, len(value)),
			Code:    "max_length",
		}
	}
	return nil
}

// MinLength validates that a string meets the minimum length.
func MinLength(field, value string, min int) *ValidationError {
	if len(value) < min {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be at least %d characters (got %d)", min, len(value)),
			Code:    "min_length",
		}
	}
	return nil
}

// Email validates that a string is a valid email address.
func Email(field, value string) *ValidationError {
	if value == "" {
		return nil // Use Required() separately if the field is mandatory
	}

	addr, err := mail.ParseAddress(value)
	if err != nil || addr.Address != value {
		return &ValidationError{
			Field:   field,
			Message: "must be a valid email address",
			Code:    "invalid_email",
		}
	}
	return nil
}

// URL validates that a string is a valid HTTP/HTTPS URL.
func URL(field, value string) *ValidationError {
	if value == "" {
		return nil // Use Required() separately if the field is mandatory
	}

	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return &ValidationError{
			Field:   field,
			Message: "must be a valid HTTP or HTTPS URL",
			Code:    "invalid_url",
		}
	}
	return nil
}

// UUID validates that a string is a valid UUID v4.
func UUID(field, value string) *ValidationError {
	if value == "" {
		return nil // Use Required() separately if the field is mandatory
	}

	normalized := strings.ToLower(value)
	if !uuidPattern.MatchString(normalized) {
		return &ValidationError{
			Field:   field,
			Message: "must be a valid UUID v4",
			Code:    "invalid_uuid",
		}
	}
	return nil
}

// ULID validates that a string is a valid ULID.
func ULID(field, value string) *ValidationError {
	if value == "" {
		return nil // Use Required() separately if the field is mandatory
	}

	if _, err := ulid.Parse(value); err != nil {
		return &ValidationError{
			Field:   field,
			Message: "must be a valid ULID",
			Code:    "invalid_ulid",
		}
	}
	return nil
}

// AlphanumericDashUnderscore validates that a string contains only alphanumeric characters, dashes, and underscores.
func AlphanumericDashUnderscore(field, value string) *ValidationError {
	if value == "" {
		return nil // Use Required() separately if the field is mandatory
	}

	if !alphanumericDashUnderscore.MatchString(value) {
		return &ValidationError{
			Field:   field,
			Message: "must contain only letters, numbers, dashes, and underscores",
			Code:    "invalid_format",
		}
	}
	return nil
}

// AlphanumericSpaceDashUnderscore validates that a string contains only alphanumeric characters, spaces, dashes, and underscores.
func AlphanumericSpaceDashUnderscore(field, value string) *ValidationError {
	if value == "" {
		return nil // Use Required() separately if the field is mandatory
	}

	if !alphanumericSpaceDashUnderscore.MatchString(value) {
		return &ValidationError{
			Field:   field,
			Message: "must contain only letters, numbers, spaces, dashes, and underscores",
			Code:    "invalid_format",
		}
	}
	return nil
}

// ModelName validates that a string is a valid AI model name.
func ModelName(field, value string) *ValidationError {
	if value == "" {
		return nil // Use Required() separately if the field is mandatory
	}

	if !modelNamePattern.MatchString(value) {
		return &ValidationError{
			Field:   field,
			Message: "must be a valid model name (e.g., 'gpt-4o', 'llama-3.1-8b')",
			Code:    "invalid_model_name",
		}
	}

	// Additional length check to prevent abuse
	if len(value) > 128 {
		return &ValidationError{
			Field:   field,
			Message: "model name is too long (max 128 characters)",
			Code:    "max_length",
		}
	}

	return nil
}

// CronExpression validates that a string is a valid cron expression.
func CronExpression(field, value string) *ValidationError {
	if value == "" {
		return nil // Use Required() separately if the field is mandatory
	}

	// Basic validation: no newlines (prevent injection), reasonable length
	if strings.ContainsAny(value, "\n\r") {
		return &ValidationError{
			Field:   field,
			Message: "must not contain newlines",
			Code:    "invalid_format",
		}
	}

	if len(value) > 100 {
		return &ValidationError{
			Field:   field,
			Message: "cron expression is too long (max 100 characters)",
			Code:    "max_length",
		}
	}

	// Validate basic cron format (minute hour day month weekday)
	if !cronPattern.MatchString(value) {
		return &ValidationError{
			Field:   field,
			Message: "must be a valid cron expression (e.g., '0 0 * * *')",
			Code:    "invalid_cron",
		}
	}

	return nil
}

// OneOf validates that a string is one of the allowed values.
func OneOf(field, value string, allowed []string) *ValidationError {
	if value == "" {
		return nil // Use Required() separately if the field is mandatory
	}

	for _, a := range allowed {
		if value == a {
			return nil
		}
	}

	return &ValidationError{
		Field:   field,
		Message: fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", ")),
		Code:    "invalid_value",
	}
}

// IntRange validates that an integer is within the specified range (inclusive).
func IntRange(field string, value, min, max int) *ValidationError {
	if value < min || value > max {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be between %d and %d (got %d)", min, max, value),
			Code:    "out_of_range",
		}
	}
	return nil
}

// PositiveInt validates that an integer is positive (> 0).
func PositiveInt(field string, value int) *ValidationError {
	if value <= 0 {
		return &ValidationError{
			Field:   field,
			Message: "must be positive",
			Code:    "must_be_positive",
		}
	}
	return nil
}

// NonNegativeInt validates that an integer is non-negative (>= 0).
func NonNegativeInt(field string, value int) *ValidationError {
	if value < 0 {
		return &ValidationError{
			Field:   field,
			Message: "must be non-negative",
			Code:    "must_be_non_negative",
		}
	}
	return nil
}

// FutureTime validates that a time is in the future.
func FutureTime(field string, value time.Time) *ValidationError {
	if value.Before(time.Now()) {
		return &ValidationError{
			Field:   field,
			Message: "must be in the future",
			Code:    "must_be_future",
		}
	}
	return nil
}

// PastTime validates that a time is in the past.
func PastTime(field string, value time.Time) *ValidationError {
	if value.After(time.Now()) {
		return &ValidationError{
			Field:   field,
			Message: "must be in the past",
			Code:    "must_be_past",
		}
	}
	return nil
}

// NoSQLInjection validates that a string doesn't contain SQL injection patterns.
// This is a defense-in-depth measure; parameterized queries are still required.
func NoSQLInjection(field, value string) *ValidationError {
	if value == "" {
		return nil
	}

	// Check for common SQL injection patterns
	dangerous := []string{
		"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_",
		"exec", "execute", "union", "select", "insert",
		"update", "delete", "drop", "create", "alter",
	}

	lowerValue := strings.ToLower(value)
	for _, pattern := range dangerous {
		if strings.Contains(lowerValue, pattern) {
			return &ValidationError{
				Field:   field,
				Message: "contains invalid characters",
				Code:    "invalid_characters",
			}
		}
	}

	return nil
}

// NoXSS validates that a string doesn't contain XSS patterns.
// This is a defense-in-depth measure; proper output encoding is still required.
func NoXSS(field, value string) *ValidationError {
	if value == "" {
		return nil
	}

	// Check for common XSS patterns
	dangerous := []string{
		"<script", "</script>", "javascript:", "onerror=", "onload=",
		"onclick=", "onmouseover=", "<iframe", "eval(", "document.cookie",
	}

	lowerValue := strings.ToLower(value)
	for _, pattern := range dangerous {
		if strings.Contains(lowerValue, pattern) {
			return &ValidationError{
				Field:   field,
				Message: "contains invalid characters",
				Code:    "invalid_characters",
			}
		}
	}

	return nil
}

// MaxSliceLength validates that a slice does not exceed the maximum length.
func MaxSliceLength[T any](field string, slice []T, max int) *ValidationError {
	if len(slice) > max {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must not contain more than %d items (got %d)", max, len(slice)),
			Code:    "max_items",
		}
	}
	return nil
}

// MinSliceLength validates that a slice meets the minimum length.
func MinSliceLength[T any](field string, slice []T, min int) *ValidationError {
	if len(slice) < min {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must contain at least %d items (got %d)", min, len(slice)),
			Code:    "min_items",
		}
	}
	return nil
}

// Validator is a helper type for chaining multiple validations.
type Validator struct {
	errors ValidationErrors
}

// NewValidator creates a new Validator instance.
func NewValidator() *Validator {
	return &Validator{
		errors: make(ValidationErrors, 0),
	}
}

// Add adds a validation error if it's not nil.
func (v *Validator) Add(err *ValidationError) *Validator {
	if err != nil {
		v.errors = append(v.errors, *err)
	}
	return v
}

// Errors returns all validation errors, or nil if there are none.
func (v *Validator) Errors() error {
	if len(v.errors) == 0 {
		return nil
	}
	return v.errors
}

// HasErrors returns true if there are any validation errors.
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}
