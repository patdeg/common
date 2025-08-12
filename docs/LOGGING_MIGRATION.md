# Add PII-Safe Logging Package

## Summary
This PR adds enhanced logging capabilities with automatic PII (Personally Identifiable Information) sanitization to the common package, supporting both AgentResume.ai and future Blue Fermion Labs applications.

## New Features

### 1. Logging Package (`logging/`)
- **`safe.go`**: Core logger with configurable log levels and output formats
- **`sanitizer.go`**: Advanced PII detection and masking engine

### 2. Enhanced Logging Functions (`logging_enhanced.go`)
- `DebugSafe()`, `InfoSafe()`, `WarnSafe()`, `ErrorSafe()` - PII-safe variants
- `SanitizeMessage()` - Manual message sanitization
- `AddCustomPIIPattern()` - Custom pattern registration
- `SetPIIProtection()` - Global PII protection toggle

## PII Patterns Detected and Masked

| Type | Example Input | Sanitized Output |
|------|--------------|------------------|
| Email | user@example.com | u***r@***.com |
| IPv4 | 192.168.1.1 | xxx.xxx.xxx.xxx |
| IPv6 | 2001:db8::1 | xxxx:xxxx:xxxx:xxxx:xxxx:xxxx:xxxx:xxxx |
| Credit Card | 4111-1111-1111-1111 | ****-****-****-1111 |
| SSN | 123-45-6789 | xxx-xx-xxxx |
| Phone | +1-555-123-4567 | +*-***-***-**** |
| API Keys | api_key=sk_live_abc123 | api_key=***REDACTED*** |
| JWT Tokens | eyJ... | ***JWT_REDACTED*** |
| Passwords | password=secret123 | password=***REDACTED*** |

## Configuration

### Environment Variables
- `LOG_PII_PROTECTION=false` - Disable PII protection (default: enabled)
- `LOG_FORMAT=json` - Output logs in JSON format
- `LOG_SOURCE=true` - Include source file/line in logs
- `DEBUG=true` or `ISDEBUG=true` - Enable debug logging

## Backward Compatibility
- ✅ Existing `Debug()`, `Info()`, `Warn()`, `Error()` functions unchanged
- ✅ New Safe variants are opt-in
- ✅ No breaking changes to existing API

## Testing
```bash
go test ./logging/...
```

## Usage Example
```go
import "github.com/patdeg/common"

// Old way (still works, but logs PII)
common.Debug("User login: email=%s", user.Email)

// New way (automatically sanitizes PII)
common.DebugSafe("User login: email=%s", user.Email)
// Output: "User login: email=u***r@***.com"
```

## Impact
- AgentResume.ai: Ready to migrate ~595 logging calls
- PNPS: Can use from day one
- Future apps: Built-in compliance from the start