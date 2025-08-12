# Package Structure and Organization

This document describes the organization and purpose of each package in the common library.

## Package Hierarchy

```
github.com/patdeg/common/
│
├── Core Utilities (Level 1 - No dependencies)
│   ├── convert.go      - Type conversion utilities
│   ├── slice.go        - Slice manipulation
│   ├── url.go          - URL utilities
│   ├── file.go         - File operations
│   └── crypt.go        - Encryption/decryption
│
├── Foundation (Level 2 - Core utilities)
│   ├── logging.go      - Basic logging
│   ├── logging_enhanced.go - PII-safe logging
│   ├── cookie.go       - Cookie handling
│   ├── debug.go        - Debug utilities
│   └── web.go          - HTTP utilities
│
├── Infrastructure (Level 3 - External services)
│   ├── auth/           - Authentication & authorization
│   ├── datastore/      - Data persistence
│   ├── bigquery/       - Analytics storage
│   ├── tasks/          - Async task processing
│   └── email/          - Email service
│
├── Business Logic (Level 4 - Domain specific)
│   ├── tenant/         - Multi-tenancy
│   ├── rbac/           - Access control
│   └── payment/        - Payment processing
│
├── Application (Level 5 - High-level features)
│   ├── frontend/       - Frontend utilities
│   ├── api/            - API client
│   ├── search/         - Search engine
│   ├── monitor/        - Health & metrics
│   └── impexp/         - Import/export
│
└── Platform Specific
    ├── gcp/            - Google Cloud Platform
    ├── ga/             - Google Analytics
    └── track/          - Custom tracking
```

## Package Descriptions

### Core Utilities

#### `convert.go`
Type conversion utilities for common Go type conversions.
- `StringToInt()`, `IntToString()`
- `StringToBool()`, `BoolToString()`
- `StringToFloat()`, `FloatToString()`
- Safe conversions with error handling

#### `slice.go`
Slice manipulation utilities for common operations.
- `Contains()` - Check if slice contains element
- `Unique()` - Remove duplicates
- `Filter()` - Filter elements
- `Map()` - Transform elements

#### `url.go`
URL manipulation and validation utilities.
- `ValidateURL()` - Validate URL format
- `GetDomain()` - Extract domain
- `AddQueryParam()` - Add query parameters
- `ParseQueryParams()` - Parse query string

#### `file.go`
File system operations and utilities.
- `FileExists()` - Check file existence
- `ReadFile()` - Safe file reading
- `WriteFile()` - Safe file writing
- `GetFileExtension()` - Extract extension

#### `crypt.go`
Encryption and hashing utilities.
- `Encrypt()` - AES-GCM encryption
- `Decrypt()` - AES-GCM decryption
- `Hash()` - SHA256 hashing
- `GenerateRandomKey()` - Key generation

### Foundation Layer

#### `logging.go` & `logging_enhanced.go`
Comprehensive logging with PII protection.
- Standard log levels (Debug, Info, Warn, Error)
- PII-safe variants (DebugSafe, InfoSafe, etc.)
- Automatic sensitive data masking
- Environment-aware (debug only in development)

#### `cookie.go`
HTTP cookie management utilities.
- `SetCookie()` - Set secure cookies
- `GetCookie()` - Retrieve cookie values
- `DeleteCookie()` - Remove cookies
- Secure defaults (HttpOnly, SameSite)

#### `debug.go`
Debug utilities for development.
- `DumpStruct()` - Pretty print structures
- `StackTrace()` - Get stack trace
- `TimedFunction()` - Measure execution time
- Development-only functions

#### `web.go`
HTTP and web utilities.
- `JSONResponse()` - Send JSON responses
- `ErrorResponse()` - Standardized error responses
- `ParseRequest()` - Request parsing
- `ValidateCSRF()` - CSRF protection

### Infrastructure Layer

#### `auth/`
Authentication and authorization system.
- OAuth2 integration (Google, GitHub)
- JWT token management
- Session handling
- User authentication middleware
- CSRF token generation

#### `datastore/`
Generic data persistence layer.
- Repository pattern implementation
- Cloud Datastore integration
- Local in-memory storage for development
- Transaction support
- Query builder

#### `bigquery/`
BigQuery analytics integration.
- Streaming inserts
- Batch processing
- Schema management
- Table creation
- Standard schemas

#### `tasks/`
Asynchronous task processing.
- Cloud Tasks integration
- Local queue for development
- Retry configuration
- Task scheduling
- Batch processing

#### `email/`
Email service abstraction.
- Multiple providers (SendGrid, SMTP)
- Template management
- Local development mode
- Attachment support
- Batch sending

### Business Logic Layer

#### `tenant/`
Multi-tenant application support.
- Tenant isolation
- Resource limits
- Subscription tiers
- Tenant context
- Domain mapping

#### `rbac/`
Role-based access control.
- Role management
- Permission system
- Policy evaluation
- User-role assignment
- Hierarchical permissions

#### `payment/`
Payment and subscription management.
- Provider abstraction (Stripe, Paddle)
- Subscription lifecycle
- Payment processing
- Webhook handling
- Usage tracking

### Application Layer

#### `frontend/`
Frontend asset and template management.
- Asset versioning
- Template rendering
- HTMX integration
- Static file serving
- Cache management

#### `api/`
HTTP client utilities.
- Retry logic
- Rate limiting
- Authentication helpers
- REST client
- Request/response handling

#### `search/`
Full-text search implementation.
- In-memory search engine
- Query builder
- Faceting
- Highlighting
- Relevance scoring

#### `monitor/`
Health monitoring and metrics.
- Health checks
- System metrics
- Custom checkers
- HTTP endpoints
- Metrics collection

#### `impexp/`
Data import/export utilities.
- Multiple formats (JSON, CSV, ZIP)
- Batch processing
- Data transformation
- Backup/restore
- Migration tools

### Platform Specific

#### `gcp/`
Google Cloud Platform utilities.
- App Engine helpers
- Datastore utilities
- User management
- Memcache wrapper
- Version detection

#### `ga/`
Google Analytics integration.
- Event tracking
- Page views
- User tracking
- E-commerce tracking
- Custom dimensions

#### `track/`
Custom tracking system.
- Event logging
- User behavior
- Custom metrics
- BigQuery integration
- Real-time analytics

## Import Guidelines

### Recommended Import Order

```go
import (
    // Standard library
    "context"
    "fmt"
    "time"
    
    // External dependencies
    "cloud.google.com/go/datastore"
    
    // Common packages (ordered by level)
    "github.com/patdeg/common"           // Core
    "github.com/patdeg/common/logging"   // Foundation
    "github.com/patdeg/common/auth"      // Infrastructure
    "github.com/patdeg/common/tenant"    // Business Logic
    "github.com/patdeg/common/frontend"  // Application
)
```

### Dependency Rules

1. **No Circular Dependencies**: Packages should only import from lower levels
2. **Minimal External Dependencies**: Prefer standard library
3. **Interface-Based**: Define interfaces for flexibility
4. **Environment Aware**: Support both development and production

## Testing Strategy

Each package should have:
- Unit tests (`*_test.go`)
- Integration tests where applicable
- Example usage in tests
- Minimum 80% coverage
- Benchmarks for performance-critical code

## Security Considerations

**REMEMBER: This is a PUBLIC repository**

- Never hardcode credentials
- Use environment variables for configuration
- Mask PII in logs
- Validate all inputs
- Use secure defaults
- Document security considerations

## Future Reorganization

### Proposed Structure (v2.0)

```
communication/
├── email/
├── sms/
├── push/
└── webhook/

storage/
├── datastore/
├── cache/
├── files/
└── search/

providers/
├── google/
├── stripe/
├── paddle/
└── sendgrid/

analytics/
├── tracking/
├── metrics/
└── reporting/
```

This reorganization would:
- Group related functionality
- Improve discoverability
- Reduce package count
- Maintain backward compatibility through aliases