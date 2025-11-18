# AGENTS.md - Guide for AI Agents

This file provides a guide for any AI agent working with the `github.com/patdeg/common` repository. Use this context to understand the project structure, conventions, and goals.

## ⚠️ CRITICAL SECURITY NOTICE

**THIS IS A PUBLIC REPOSITORY.** Under no circumstances should you ever commit:
- API keys, tokens, or secrets.
- Personal information (PII) of any kind.
- Credentials, passwords, or auth tokens.
- Any sensitive data.

All configuration and secrets must be loaded from environment variables. Use placeholders and example domains (`example.com`) in the code.

## Project Overview

This repository, `common`, is a Go library containing shared utility packages. Its purpose is to provide a centralized location for code that is reused across multiple web applications and services, promoting consistency and reducing duplication.

### Core Principles
1.  **Security**: The codebase must remain secure and free of sensitive data.
2.  **Reliability**: Code should be robust and well-tested.
3.  **Clarity**: Code must be well-documented and easy to understand.
4.  **Compatibility**: Changes should not break existing functionality.

## Package Organization

The library is organized into packages based on functionality:

*   **Core Utilities (`/`):** Foundational helpers for tasks like type conversion (`convert`), cryptography (`crypt`), and logging (`logging`).
*   **Web & API (`/`):** HTTP-related utilities, including security middleware (`web`), cookie management (`cookie`), and CSRF protection (`csrf`).
*   **GCP Integration (`/gcp`, `/datastore`, `/bigquery`):** Helpers for interacting with Google Cloud Platform services like Datastore and BigQuery.
*   **Application Features (`/auth`, `/email`, etc.):** Components for higher-level functionality such as user authentication and sending emails.
*   **Analytics (`/ga`, `/track`):** Tools for Google Analytics and custom event tracking.

## Building and Running

### Installation

To use this library in another Go project, run:
```bash
go get github.com/patdeg/common
```

### Testing Commands

The project has a suite of tests. Use the following commands to validate changes:

```bash
# Run all tests in the repository
go test ./...

# Run tests with the race detector to find concurrency issues
go test -race ./...

# Calculate and view code coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Running Examples

The `examples/` directory shows how to use the library. To run an example:
```bash
cd examples/basic-usage
go run main.go
```

## Key Development Guidelines

### Code Style
- Follow standard Go formatting (`gofmt`).
- All exported functions, types, and variables must have clear documentation comments.

### Error Handling
- Functions should return errors instead of panicking.
- Provide context when wrapping errors (e.g., `fmt.Errorf("doing X failed: %w", err)`).

### Environment Variables
Configuration and secrets must be loaded from the environment. Do not hardcode them.

```go
// Correct: Load from environment.
apiKey := os.Getenv("API_KEY")
if apiKey == "" {
    log.Fatal("API_KEY environment variable not set")
}

// Incorrect: Hardcoded secret. Do not do this.
apiKey := "pa_12345..."
```

### Dependencies
- Keep external dependencies to a minimum.
- Lower-level packages should not import higher-level packages.

## Security Checklist

Before committing, always double-check:
- [ ] Is the code free of any credentials or secrets?
- [ ] Is there any personal or user data in the changes?
- [ ] Are all sensitive values loaded from the environment?