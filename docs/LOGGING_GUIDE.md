# Logging Guide

## Overview

The common package provides comprehensive logging capabilities with automatic PII (Personally Identifiable Information) protection. This guide covers both the standard logging functions and the enhanced PII-safe variants.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Standard Logging](#standard-logging)
3. [PII-Safe Logging](#pii-safe-logging)
4. [Configuration](#configuration)
5. [PII Detection Patterns](#pii-detection-patterns)
6. [Migration Guide](#migration-guide)
7. [Best Practices](#best-practices)

## Quick Start

```go
import "github.com/patdeg/common"

// Standard logging (use for non-sensitive data)
common.Info("Server started on port %d", 8080)
common.Error("Failed to connect to database")

// PII-safe logging (use when data might contain PII)
common.InfoSafe("User logged in: %s", user.Email)  // Email will be masked
common.ErrorSafe("Failed to process payment for user %s", user.Email)
```

## Standard Logging

The package provides four standard logging levels:

### Debug
```go
common.Debug("Debug message: %v", data)
```
- Only outputs when `ISDEBUG` environment variable is `true`
- Use for detailed diagnostic information

### Info
```go
common.Info("Informational message: %s", status)
```
- General informational messages
- Always outputs

### Warn
```go
common.Warn("Warning message: %s", issue)
```
- Warning conditions that should be reviewed
- Prefixed with "WARNING:" for easy grep

### Error
```go
common.Error("Error message: %v", err)
```
- Error conditions requiring attention
- Prefixed with "ERROR:" for easy grep

## PII-Safe Logging

The Safe variants automatically detect and mask PII before logging:

### DebugSafe
```go
common.DebugSafe("User activity: email=%s, ip=%s", email, ipAddr)
// Output: "User activity: email=u***@***.com, ip=xxx.xxx.xxx.xxx"
```

### InfoSafe
```go
common.InfoSafe("Processing payment for card %s", cardNumber)
// Output: "Processing payment for card ****-****-****-1234"
```

### WarnSafe
```go
common.WarnSafe("Invalid SSN format: %s", ssn)
// Output: "Invalid SSN format: xxx-xx-xxxx"
```

### ErrorSafe
```go
common.ErrorSafe("API call failed with key: %s", apiKey)
// Output: "API call failed with key: ***REDACTED***"
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ISDEBUG` or `DEBUG` | Enable debug logging | `false` |
| `LOG_PII_PROTECTION` | Enable PII sanitization | `true` |
| `LOG_FORMAT` | Output format (`text` or `json`) | `text` |
| `LOG_SOURCE` | Include source file/line in logs | `false` |

### Programmatic Configuration

```go
// Disable PII protection globally
common.SetPIIProtection(false)

// Add custom PII pattern
common.AddCustomPIIPattern("employee_id", `\bEMP\d{6}\b`)

// Manually sanitize a message
sanitized := common.SanitizeMessage("User email: john@example.com")
```

## PII Detection Patterns

The sanitizer automatically detects and masks these patterns:

### Personal Identifiers

| Type | Example | Masked Output |
|------|---------|---------------|
| Email | `john.doe@company.com` | `j***e@***.com` |
| SSN | `123-45-6789` | `xxx-xx-xxxx` |
| Phone | `+1-555-123-4567` | `+*-***-***-****` |

### Network Information

| Type | Example | Masked Output |
|------|---------|---------------|
| IPv4 | `192.168.1.100` | `xxx.xxx.xxx.xxx` |
| IPv6 | `2001:db8::8a2e:370:7334` | `xxxx:xxxx:xxxx:xxxx:xxxx:xxxx:xxxx:xxxx` |

### Financial Data

| Type | Example | Masked Output |
|------|---------|---------------|
| Credit Card | `4111-1111-1111-1111` | `****-****-****-1111` |
| Credit Card (spaces) | `4111 1111 1111 1111` | `****-****-****-1111` |

### Authentication & Security

| Type | Example | Masked Output |
|------|---------|---------------|
| API Key | `api_key=sk_live_abc123def` | `api_key=***REDACTED***` |
| JWT Token | `eyJhbGciOiJIUzI1NiIs...` | `***JWT_REDACTED***` |
| Password | `password=MySecret123!` | `password=***REDACTED***` |

## Migration Guide

### Step 1: Identify PII Logging

Search for logging calls that include PII:
```bash
# Find potential PII logging
grep -r "common\.\(Debug\|Info\|Warn\|Error\).*email" .
grep -r "common\.\(Debug\|Info\|Warn\|Error\).*password" .
grep -r "common\.\(Debug\|Info\|Warn\|Error\).*token" .
```

### Step 2: Update Critical Paths

Start with authentication and user data:
```go
// Before
common.Info("User login: %s", user.Email)
common.Debug("Token generated: %s", token)

// After
common.InfoSafe("User login: %s", user.Email)
common.DebugSafe("Token generated: %s", token)
```

### Step 3: Gradual Migration

You don't need to migrate everything at once:
1. Start with new code - use Safe variants by default
2. Update high-risk areas (auth, payments, user data)
3. Gradually migrate other areas during refactoring

### Step 4: Verification

Test that PII is being masked:
```go
// Add temporary test
common.InfoSafe("TEST: email@example.com, 192.168.1.1, 4111-1111-1111-1111")
// Should output: "TEST: e***l@***.com, xxx.xxx.xxx.xxx, ****-****-****-1111"
```

## Best Practices

### 1. Use Safe Variants for User Data

Always use Safe variants when logging user-provided data:
```go
// Good
common.InfoSafe("User registered: %s", user.Email)

// Bad - exposes PII
common.Info("User registered: %s", user.Email)
```

### 2. Log Events, Not Data

When possible, log what happened rather than the data:
```go
// Better
common.Info("User registration successful")

// Good if you need the email
common.InfoSafe("User registered: %s", user.Email)
```

### 3. Use Debug for Detailed Logs

Put PII-heavy logs at Debug level so they're disabled in production:
```go
common.DebugSafe("Full user object: %+v", user)
```

### 4. Structured Logging

For complex data, consider structured logging:
```go
// The sanitizer will process struct fields
userMap := common.SanitizeStruct(user)
common.Info("User data: %v", userMap)
```

### 5. Custom Patterns

Add domain-specific patterns for your application:
```go
func init() {
    // Mask employee IDs
    common.AddCustomPIIPattern("employee_id", `\bEMP\d{6}\b`)
    
    // Mask internal user IDs
    common.AddCustomPIIPattern("user_id", `\bUSR[A-Z0-9]{8}\b`)
}
```

### 6. Testing

Always test that PII is being properly masked:
```go
func TestPIIMasking(t *testing.T) {
    // Capture log output
    var buf bytes.Buffer
    log.SetOutput(&buf)
    
    // Log with PII
    common.InfoSafe("Test: john@example.com")
    
    // Verify masking
    output := buf.String()
    if strings.Contains(output, "john@example.com") {
        t.Error("Email was not masked")
    }
    if !strings.Contains(output, "j***n@***.com") {
        t.Error("Email was not properly masked")
    }
}
```

## Debugging

### Check if PII Protection is Active

```go
if common.EnablePIIProtection {
    fmt.Println("PII protection is enabled")
}
```

### Test Sanitization

```go
// Test what will be logged
msg := fmt.Sprintf("User %s logged in from %s", email, ip)
sanitized := common.SanitizeMessage(msg)
fmt.Printf("Original: %s\nSanitized: %s\n", msg, sanitized)
```

### Enable Source Location

For debugging, enable source file/line in logs:
```bash
export LOG_SOURCE=true
```

Output will include source:
```
[INFO] User logged in (auth/login.go:45)
```

## Performance Considerations

- PII sanitization adds minimal overhead (~1-5 microseconds per log)
- Regex patterns are compiled once and cached
- Safe variants only apply sanitization when PII protection is enabled
- In performance-critical paths, you can disable protection:
  ```go
  common.SetPIIProtection(false)
  // Critical section
  common.SetPIIProtection(true)
  ```

## Compliance

Using PII-safe logging helps with:
- **GDPR**: Prevents unnecessary PII in logs
- **CCPA**: Reduces PII exposure
- **HIPAA**: Helps protect health information
- **PCI DSS**: Masks credit card numbers
- **Internal Policies**: Enforces consistent PII handling

## Troubleshooting

### PII Not Being Masked

1. Check if protection is enabled:
   ```go
   fmt.Println("Protection enabled:", common.EnablePIIProtection)
   ```

2. Verify you're using Safe variants:
   ```go
   common.InfoSafe("...")  // Correct
   common.Info("...")      // Won't mask PII
   ```

3. Test the pattern directly:
   ```go
   result := common.SanitizeMessage("test@example.com")
   fmt.Println(result) // Should show masked email
   ```

### Custom Patterns Not Working

```go
// Check pattern compilation
err := common.AddCustomPIIPattern("test", `invalid(pattern`)
if err != nil {
    fmt.Printf("Pattern error: %v\n", err)
}
```

## Support

For issues or questions:
1. Check this guide
2. Review test files in `logging/` package
3. Open an issue on GitHub