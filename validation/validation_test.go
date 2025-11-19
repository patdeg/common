package validation

import (
	"testing"
	"time"
)

func TestRequired(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		wantError bool
	}{
		{"non-empty value", "username", "john", false},
		{"empty string", "username", "", true},
		{"whitespace only", "username", "   ", true},
		{"tab characters", "username", "\t\t", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Required(tt.field, tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("Required() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestMaxLength(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		max       int
		wantError bool
	}{
		{"within limit", "hello", 10, false},
		{"exactly at limit", "hello", 5, false},
		{"exceeds limit", "hello world", 5, true},
		{"empty string", "", 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MaxLength("field", tt.value, tt.max)
			if (err != nil) != tt.wantError {
				t.Errorf("MaxLength() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestMinLength(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		min       int
		wantError bool
	}{
		{"meets minimum", "hello", 3, false},
		{"exactly at minimum", "hello", 5, false},
		{"below minimum", "hi", 5, true},
		{"empty string", "", 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MinLength("field", tt.value, tt.min)
			if (err != nil) != tt.wantError {
				t.Errorf("MinLength() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestEmail(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"valid email", "user@example.com", false},
		{"valid email with subdomain", "user@mail.example.com", false},
		{"valid email with plus", "user+tag@example.com", false},
		{"invalid - no @", "userexample.com", true},
		{"invalid - no domain", "user@", true},
		{"invalid - no user", "@example.com", true},
		{"invalid - spaces", "user @example.com", true},
		{"empty string (allowed)", "", false}, // Use Required() separately
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Email("email", tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("Email() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestURL(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"valid http URL", "http://example.com", false},
		{"valid https URL", "https://example.com", false},
		{"valid URL with path", "https://example.com/path/to/resource", false},
		{"valid URL with query", "https://example.com?foo=bar", false},
		{"invalid - no scheme", "example.com", true},
		{"invalid - ftp scheme", "ftp://example.com", true},
		{"invalid - no host", "https://", true},
		{"invalid - malformed", "https:/example.com", true},
		{"empty string (allowed)", "", false}, // Use Required() separately
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := URL("url", tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("URL() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestUUID(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"valid UUID v4", "550e8400-e29b-41d4-a716-446655440000", false},
		{"valid UUID v4 uppercase", "550E8400-E29B-41D4-A716-446655440000", false},
		{"invalid - wrong format", "not-a-uuid", true},
		{"invalid - missing dashes", "550e8400e29b41d4a716446655440000", true},
		{"invalid - wrong version", "550e8400-e29b-31d4-a716-446655440000", true},
		{"empty string (allowed)", "", false}, // Use Required() separately
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := UUID("uuid", tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("UUID() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestULID(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"valid ULID", "01ARZ3NDEKTSV4RRFFQ69G5FAV", false},
		{"invalid - wrong length", "01ARZ3NDEKTSV4RRFFQ", true},
		{"invalid - not a ULID", "not-a-valid-ulid-at-all-123", true},
		{"empty string (allowed)", "", false}, // Use Required() separately
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ULID("ulid", tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ULID() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestAlphanumericDashUnderscore(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"valid alphanumeric", "abc123", false},
		{"valid with dash", "abc-123", false},
		{"valid with underscore", "abc_123", false},
		{"valid mixed", "abc-123_xyz", false},
		{"invalid - space", "abc 123", true},
		{"invalid - special char", "abc@123", true},
		{"empty string (allowed)", "", false}, // Use Required() separately
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AlphanumericDashUnderscore("field", tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("AlphanumericDashUnderscore() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestAlphanumericSpaceDashUnderscore(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"valid with space", "abc 123", false},
		{"valid with dash", "abc-123", false},
		{"valid with underscore", "abc_123", false},
		{"valid mixed", "abc 123-xyz_foo", false},
		{"invalid - special char", "abc@123", true},
		{"invalid - newline", "abc\n123", true},
		{"empty string (allowed)", "", false}, // Use Required() separately
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AlphanumericSpaceDashUnderscore("field", tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("AlphanumericSpaceDashUnderscore() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestModelName(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"valid OpenAI model", "gpt-4o", false},
		{"valid Anthropic model", "claude-3-opus-20240229", false},
		{"valid Groq model", "llama-3.1-8b-instant", false},
		{"valid with path", "openai/gpt-4o", false},
		{"valid with underscores", "model_name_v2", false},
		{"invalid - too long", "this-is-a-very-long-model-name-that-exceeds-the-maximum-allowed-length-of-128-characters-and-should-be-rejected-by-validation-logic-to-prevent-abuse", true},
		{"invalid - special chars", "model@name", true},
		{"invalid - spaces", "model name", true},
		{"empty string (allowed)", "", false}, // Use Required() separately
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ModelName("model", tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("ModelName() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestCronExpression(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"valid - every hour", "0 * * * *", false},
		{"valid - every day at midnight", "0 0 * * *", false},
		{"valid - every 15 minutes", "*/15 * * * *", false},
		{"valid - specific hour", "0 12 * * 0", false},
		{"invalid - newline", "0 0 * * *\n", true},
		{"invalid - too long", "0 0 * * * 0 0 * * * 0 0 * * * 0 0 * * * 0 0 * * * 0 0 * * * 0 0 * * * 0 0 * * * 0 0 * * * 0 0 * * *", true},
		{"invalid - malformed", "not a cron", true},
		{"empty string (allowed)", "", false}, // Use Required() separately
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CronExpression("schedule", tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("CronExpression() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestOneOf(t *testing.T) {
	allowed := []string{"apple", "banana", "cherry"}

	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"valid - first", "apple", false},
		{"valid - middle", "banana", false},
		{"valid - last", "cherry", false},
		{"invalid - not in list", "orange", true},
		{"invalid - case sensitive", "Apple", true},
		{"empty string (allowed)", "", false}, // Use Required() separately
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := OneOf("fruit", tt.value, allowed)
			if (err != nil) != tt.wantError {
				t.Errorf("OneOf() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestIntRange(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		min       int
		max       int
		wantError bool
	}{
		{"within range", 5, 1, 10, false},
		{"at minimum", 1, 1, 10, false},
		{"at maximum", 10, 1, 10, false},
		{"below minimum", 0, 1, 10, true},
		{"above maximum", 11, 1, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IntRange("value", tt.value, tt.min, tt.max)
			if (err != nil) != tt.wantError {
				t.Errorf("IntRange() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestPositiveInt(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		wantError bool
	}{
		{"positive", 1, false},
		{"large positive", 1000, false},
		{"zero", 0, true},
		{"negative", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PositiveInt("value", tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("PositiveInt() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestNonNegativeInt(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		wantError bool
	}{
		{"positive", 1, false},
		{"zero", 0, false},
		{"negative", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NonNegativeInt("value", tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("NonNegativeInt() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestFutureTime(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		value     time.Time
		wantError bool
	}{
		{"one hour in future", now.Add(time.Hour), false},
		{"one day in future", now.Add(24 * time.Hour), false},
		{"one hour in past", now.Add(-time.Hour), true},
		{"one day in past", now.Add(-24 * time.Hour), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := FutureTime("expires_at", tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("FutureTime() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestPastTime(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		value     time.Time
		wantError bool
	}{
		{"one hour in past", now.Add(-time.Hour), false},
		{"one day in past", now.Add(-24 * time.Hour), false},
		{"one hour in future", now.Add(time.Hour), true},
		{"one day in future", now.Add(24 * time.Hour), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := PastTime("created_at", tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("PastTime() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestNoSQLInjection(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"safe string", "hello world", false},
		{"safe with numbers", "user123", false},
		{"dangerous - single quote", "user'123", true},
		{"dangerous - double quote", "user\"123", true},
		{"dangerous - semicolon", "user;123", true},
		{"dangerous - comment", "user--123", true},
		{"dangerous - union", "1 UNION SELECT", true},
		{"dangerous - exec", "EXEC sp_", true},
		{"empty string (allowed)", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NoSQLInjection("field", tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("NoSQLInjection() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestNoXSS(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"safe string", "hello world", false},
		{"safe with numbers", "user123", false},
		{"dangerous - script tag", "<script>alert(1)</script>", true},
		{"dangerous - javascript protocol", "javascript:alert(1)", true},
		{"dangerous - onerror", "img onerror=alert(1)", true},
		{"dangerous - onload", "body onload=alert(1)", true},
		{"dangerous - iframe", "<iframe src=evil.com>", true},
		{"dangerous - eval", "eval(document.cookie)", true},
		{"empty string (allowed)", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NoXSS("field", tt.value)
			if (err != nil) != tt.wantError {
				t.Errorf("NoXSS() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestMaxSliceLength(t *testing.T) {
	tests := []struct {
		name      string
		slice     []string
		max       int
		wantError bool
	}{
		{"within limit", []string{"a", "b", "c"}, 5, false},
		{"exactly at limit", []string{"a", "b", "c"}, 3, false},
		{"exceeds limit", []string{"a", "b", "c", "d"}, 3, true},
		{"empty slice", []string{}, 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MaxSliceLength("items", tt.slice, tt.max)
			if (err != nil) != tt.wantError {
				t.Errorf("MaxSliceLength() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestMinSliceLength(t *testing.T) {
	tests := []struct {
		name      string
		slice     []string
		min       int
		wantError bool
	}{
		{"meets minimum", []string{"a", "b", "c"}, 2, false},
		{"exactly at minimum", []string{"a", "b", "c"}, 3, false},
		{"below minimum", []string{"a"}, 3, true},
		{"empty slice", []string{}, 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MinSliceLength("items", tt.slice, tt.min)
			if (err != nil) != tt.wantError {
				t.Errorf("MinSliceLength() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidator(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		v := NewValidator()
		v.Add(Required("username", "john"))
		v.Add(Email("email", "john@example.com"))

		if v.HasErrors() {
			t.Errorf("Validator.HasErrors() = true, want false")
		}
		if v.Errors() != nil {
			t.Errorf("Validator.Errors() = %v, want nil", v.Errors())
		}
	})

	t.Run("single error", func(t *testing.T) {
		v := NewValidator()
		v.Add(Required("username", ""))

		if !v.HasErrors() {
			t.Errorf("Validator.HasErrors() = false, want true")
		}
		if v.Errors() == nil {
			t.Errorf("Validator.Errors() = nil, want error")
		}
	})

	t.Run("multiple errors", func(t *testing.T) {
		v := NewValidator()
		v.Add(Required("username", ""))
		v.Add(Email("email", "invalid"))
		v.Add(URL("website", "not-a-url"))

		if !v.HasErrors() {
			t.Errorf("Validator.HasErrors() = false, want true")
		}

		err := v.Errors()
		if err == nil {
			t.Errorf("Validator.Errors() = nil, want error")
		}

		// Check that error message contains information about multiple failures
		errMsg := err.Error()
		if !contains(errMsg, "username") || !contains(errMsg, "email") || !contains(errMsg, "website") {
			t.Errorf("Validator.Errors() message doesn't contain all field names: %s", errMsg)
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexString(s, substr) >= 0)
}

func indexString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
