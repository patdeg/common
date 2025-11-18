# CLAUDE.md - Common Package Development Guide

This file provides guidance to Claude when working with the `github.com/patdeg/common` repository. Follow these instructions to ensure code quality, security, and consistency.

## ⚠️ CRITICAL SECURITY NOTICE

**THIS IS A PUBLIC REPOSITORY.** Never commit:
- API keys, tokens, or secrets (use environment variables).
- Email addresses (except `example.com`).
- Personal information or PII.
- Credentials of any kind.
- Customer data or real user information.
- Internal URLs or sensitive endpoints.

Always use environment variables for configuration, example domains (`example.com`), and generic placeholders for sensitive data.

## Project Overview

The `common` package is a shared Go library providing reusable components for multiple applications. It serves as a centralized repository for common functionality to avoid code duplication and maintain consistency.

### Core Principles
1.  **Security First**: This is a PUBLIC repository; no sensitive data.
2.  **Interface-Based Design**: Define interfaces for flexibility and testability.
3.  **Environment Awareness**: Automatically detect development vs. production environments.
4.  **Minimal Dependencies**: Keep external dependencies to a minimum.
5.  **Backward Compatibility**: Do not break existing APIs.
6.  **Comprehensive Testing**: Every package must have thorough tests.
7.  **Clear Documentation**: Every public function must be documented.

## Package Organization

The library is organized into packages based on functionality:

*   **Core Utilities (`/`):** `common`, `convert`, `crypt`, `file`, `slice`, `url`. These provide foundational helpers for logging, type conversion, and more.
*   **Web & API (`/`):** `web`, `cookie`, `csrf`. These handle HTTP-level concerns like middleware, security, and bot detection.
*   **GCP Integration (`/gcp`, `/datastore`, `/bigquery`):** A repository pattern for Datastore, BigQuery utilities, and other Google Cloud Platform helpers.
*   **Application Features (`/auth`, `/email`, `/payment`, `/search`, `/monitor`):** Higher-level features like authentication, email, payments, search, and monitoring.
*   **Analytics (`/ga`, `/track`):** Google Analytics and custom event tracking.

## Building and Running

### Testing Commands
Use the following commands to test the codebase:

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./auth

# Run with race detection
go test -race ./...

# Generate a coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Running Examples
The `examples/` directory contains runnable examples:
```bash
# Example for basic usage
cd examples/basic-usage && go run main.go

# Example for a full App Engine application
cd examples/appengine && go run main.go
```

## Development Guidelines

### Adding New Packages
1.  Create a descriptive, singular, lowercase package name.
2.  Add a `doc.go` file with package documentation.
3.  Define interfaces before implementations.
4.  Provide both cloud and local/mock implementations.
5.  Add comprehensive, table-driven tests.
6.  Document all exported functions.

### Code Style
```go
// Package packagename provides a clear, one-sentence summary.
package packagename

// InterfaceName defines what the component does.
type InterfaceName interface {
    // MethodName does something specific and returns an error on failure.
    MethodName(ctx context.Context, param string) error
}

// structName implements InterfaceName. It is unexported if not needed outside the package.
type structName struct {
    // Exported fields have JSON tags.
    Field string `json:"field"`
    
    // Unexported fields are for internal state.
    mu sync.RWMutex
}

// New creates a new instance of the component.
// Always validate inputs and set sensible defaults.
func New(config Config) (*StructName, error) {
    if config.Field == "" {
        config.Field = "default"
    }
    return &StructName{
        Field: config.Field,
    }, nil
}
```

### Testing Requirements
1.  Test coverage should be >80%.
2.  Use table-driven tests for clarity.
3.  Test both success and error cases thoroughly.
4.  Mock external dependencies using interfaces.

### Environment Variables
Never hardcode sensitive values. Use environment variables.

```go
// Good
apiKey := os.Getenv("SENDGRID_API_KEY")
if apiKey == "" {
    return errors.New("SENDGRID_API_KEY not set")
}

// Bad
apiKey := "sk_live_abc123" // NEVER DO THIS
```

### Error Handling
1.  Always return errors; do not panic.
2.  Wrap errors with `fmt.Errorf("context: %w", err)` to add context.
3.  Use sentinel errors (`var ErrNotFound = errors.New("not found")`) for known conditions and check with `errors.Is`.

## Package Dependencies

- **Level 1 (No dependencies):** `convert`, `slice`, `url`, `file`, `crypt`
- **Level 2 (Core utilities):** `logging`, `cookie`, `debug`
- **Level 3 (Infrastructure):** `auth`, `datastore`, `bigquery`, `tasks`, `email`
- **Level 4 (Business logic):** `tenant`, `rbac`, `payment`
- **Level 5 (Application):** `frontend`, `api`, `monitor`, `search`

**Import Rules:**
1.  Standard library first.
2.  External dependencies second.
3.  Internal packages last.
4.  Alphabetically ordered within groups.

## Security Checklist

Before committing, ensure:
- [ ] No hardcoded credentials.
- [ ] No real email addresses (use `example.com`).
- [ ] No customer or user data.
- [ ] All secrets are loaded from environment variables.
- [ ] Proper input validation is in place.
- [ ] Defenses against SQL injection and XSS are used where applicable.
