# CLAUDE.md - Common Package Development Guide

This file provides guidance to Claude Code (claude.ai/code) when working with the `github.com/patdeg/common` repository.

## ⚠️ CRITICAL SECURITY NOTICE

**THIS IS A PUBLIC REPOSITORY** - Never commit:
- API keys, tokens, or secrets (use environment variables)
- Email addresses (except example.com)
- Personal information or PII
- Credentials of any kind
- Customer data or real user information
- Internal URLs or endpoints
- Sensitive business logic

Always use:
- Environment variables for configuration
- Example domains (example.com, example.org)
- Generic placeholders for sensitive data
- Proper .gitignore patterns

## Repository Overview

The `common` package is a shared Go library providing reusable components for multiple applications. It serves as a centralized repository for common functionality to avoid code duplication across projects.

### Core Principles

1. **Security First**: This is a PUBLIC repository - no sensitive data
2. **Interface-Based Design**: Define interfaces for flexibility
3. **Environment Awareness**: Auto-detect development vs production
4. **Zero Dependencies**: Minimize external dependencies where possible
5. **Backward Compatibility**: Never break existing APIs
6. **Comprehensive Testing**: Every package must have tests
7. **Clear Documentation**: Every public function must be documented

## Package Organization

### Core Utilities (`/`)
- `common.go` - Core constants and initialization
- `convert.go` - Type conversion utilities
- `cookie.go` - Cookie handling
- `crypt.go` - Encryption/decryption utilities
- `debug.go` - Debug utilities
- `file.go` - File operations
- `interfaces.go` - Common interfaces
- `logging.go` - Basic logging
- `logging_enhanced.go` - PII-safe logging
- `slice.go` - Slice operations
- `url.go` - URL utilities
- `web.go` - Web/HTTP utilities

### Authentication & Security (`/auth`)
- OAuth2 integration (Google, GitHub)
- JWT token handling
- Session management
- CSRF protection

### Data Storage (`/datastore`)
- Generic repository pattern
- Cloud Datastore integration
- Local in-memory storage for development
- Transaction support

### Analytics (`/bigquery`)
- BigQuery client with batch processing
- Schema management
- Streaming inserts
- Standard schemas for common use cases

### Task Processing (`/tasks`)
- Cloud Tasks integration
- Local queue for development
- Retry configuration
- Batch processing

### Communication

#### Email (`/email`)
- SendGrid integration
- SMTP support
- Template management
- Local development mode

#### SMS/Voice (`/twilio`) - *TODO: Consolidate with communication*
- SMS sending
- Voice calls
- WhatsApp integration

### Business Logic

#### Multi-tenancy (`/tenant`)
- Tenant isolation
- Limit management
- Subscription tiers

#### Access Control (`/rbac`)
- Role-based access control
- Permission management
- Policy evaluation

#### Payments (`/payment`)
- Subscription management
- Payment processing
- Provider abstraction (Stripe, Paddle)
- Webhook handling

### Frontend Support (`/frontend`)
- Asset management with versioning
- Template rendering
- HTMX helpers
- Static file serving

### Search (`/search`)
- Full-text search
- In-memory engine
- Faceting and highlighting
- Query builder

### Operations

#### Monitoring (`/monitor`)
- Health checks
- Metrics collection
- System monitoring
- HTTP endpoints

#### Import/Export (`/impexp`)
- Data migration
- Multiple format support (JSON, CSV, ZIP)
- Batch processing
- Backup/restore

### API Integration (`/api`)
- HTTP client with retry logic
- Rate limiting
- Authentication helpers
- REST client utilities

### Analytics & Tracking

#### Google Analytics (`/ga`)
- Event tracking
- User analytics
- Conversion tracking

#### Custom Tracking (`/track`)
- Internal event tracking
- User behavior analytics

### Google Cloud Platform (`/gcp`)
- App Engine helpers
- Datastore utilities
- BigQuery integration
- Memcache wrapper
- User management

## Development Guidelines

### Adding New Packages

1. Create a descriptive package name (singular, lowercase)
2. Add a `doc.go` file with package documentation
3. Define interfaces before implementations
4. Provide both cloud and local implementations
5. Add comprehensive tests
6. Document all exported functions

### Code Style

```go
// Package description should be comprehensive
package packagename

// InterfaceName defines what the component does
type InterfaceName interface {
    // MethodName does something specific
    // Returns error if something goes wrong
    MethodName(ctx context.Context, param string) error
}

// StructName implements InterfaceName
type StructName struct {
    // exported fields with json tags
    Field string `json:"field"`
    
    // unexported fields for internal use
    mu sync.RWMutex
}

// NewStructName creates a new instance
// Always validate inputs and set defaults
func NewStructName(config Config) (*StructName, error) {
    if config.Field == "" {
        config.Field = "default"
    }
    return &StructName{
        Field: config.Field,
    }, nil
}
```

### Testing Requirements

1. Test coverage should be >80%
2. Use table-driven tests
3. Test both success and error cases
4. Mock external dependencies
5. Provide integration tests where applicable

### Environment Variables

Never hardcode sensitive values. Use environment variables:

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

1. Always return errors, don't panic
2. Wrap errors with context
3. Use sentinel errors for known conditions
4. Log errors appropriately (without PII)

```go
var ErrNotFound = errors.New("not found")

func GetItem(id string) (*Item, error) {
    item, err := db.Get(id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrNotFound
        }
        return nil, fmt.Errorf("failed to get item %s: %w", id, err)
    }
    return item, nil
}
```

## Package Dependencies

### Recommended Structure

```
Level 1 (No dependencies):
- convert, slice, url, file, crypt

Level 2 (Core utilities):
- logging, cookie, debug

Level 3 (Infrastructure):
- auth, datastore, bigquery, tasks, email

Level 4 (Business logic):
- tenant, rbac, payment

Level 5 (Application):
- frontend, api, monitor, search
```

### Import Rules

1. Standard library first
2. External dependencies second
3. Internal packages last
4. Alphabetically ordered within groups

## Security Checklist

Before committing:

- [ ] No hardcoded credentials
- [ ] No real email addresses (use example.com)
- [ ] No API endpoints with real domains
- [ ] No customer or user data
- [ ] No internal business logic details
- [ ] All secrets in environment variables
- [ ] Proper input validation
- [ ] SQL injection prevention
- [ ] XSS prevention in web utilities

## Reorganization Plan

### TODO: Package Consolidation

1. **Communication Package** (`/comm`)
   - Move `/email` here
   - Move `/twilio` here (SMS/Voice)
   - Add push notifications
   - Add webhook utilities

2. **Storage Package** (`/storage`)
   - Keep `/datastore` as is
   - Move GCS utilities here
   - Add cache abstraction
   - Add file storage abstraction

3. **Analytics Package** (`/analytics`)
   - Merge `/ga` and `/track`
   - Add `/bigquery` analytics helpers
   - Add metrics aggregation

4. **Provider Package** (`/providers`)
   - Move payment providers here
   - Add OAuth providers
   - Add cloud providers
   - Organize by company (stripe/, paddle/, google/)

## Testing Commands

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./auth

# Run with race detection
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Release Process

1. Ensure all tests pass
2. Update version tags
3. Document breaking changes
4. Create GitHub release
5. Update dependent projects

## Common Issues

### Issue: Import cycles
**Solution**: Review package dependencies, move shared types to interfaces package

### Issue: Missing environment variables
**Solution**: Provide defaults or clear error messages

### Issue: PII in logs
**Solution**: Use PII-safe logging functions from logging package

## Communication

When work is completed or requires user notification:

```bash
~/bin/sendphone.sh "Your completion message here"
```

This command pings the user with status updates about important work completion.

## Contributing

1. Create feature branch
2. Add tests for new functionality
3. Ensure no sensitive data
4. Run security checks
5. Submit PR with clear description

## Future Enhancements

- [ ] Add OpenTelemetry support
- [ ] Add distributed tracing
- [ ] Add circuit breaker patterns
- [ ] Add feature flags system
- [ ] Add A/B testing utilities
- [ ] Add GraphQL support
- [ ] Add WebSocket utilities
- [ ] Add job queue system
- [ ] Add workflow engine
- [ ] Add rule engine