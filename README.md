# Common Go Package

[![Go Reference](https://pkg.go.dev/badge/github.com/patdeg/common.svg)](https://pkg.go.dev/github.com/patdeg/common)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

A comprehensive Go library providing reusable components for building scalable applications. This package offers production-ready utilities for authentication, data storage, payments, monitoring, and more.

## 🚀 Features

- **🧾 Logging**: Structured, PII-safe logging with optional LLM-assisted error analysis
- **🔐 Authentication**: OAuth2, JWT, session management
- **💾 Data Storage**: Datastore, BigQuery, caching abstractions
- **💰 Payments**: Subscription management with multiple providers
- **📧 Communications**: Email, SMS, and notification services
- **🔍 Search**: Full-text search with faceting
- **📊 Analytics**: Event tracking and metrics collection
- **🏥 Monitoring**: Health checks and system metrics
- **🎨 Frontend**: Asset management, templates, HTMX support
- **🔒 Security**: RBAC, multi-tenancy, PII protection

## 📦 Installation

```bash
go get github.com/patdeg/common
```

## 🛠️ Quick Start

### Basic Usage
```go
import "github.com/patdeg/common"

// PII-safe logging
common.InfoSafe("User logged in: %s", userEmail)

// Type conversions
num, err := common.StringToInt("123")
```

### LLM-Assisted Logging
```go
log := common.CreateLoggingLLM("handler.go", "HandleRequest", "processing %s", r.URL.Path)
defer log.Print() // emit markdown summary at the end

log.InfoSafe("starting handler for user %s", userEmail)

if err := doWork(); err != nil {
    // Triggers asynchronous LLM analysis when COMMON_LLM_API_KEY is set
    log.ErrorSafe("failed to process request: %v", err)
    return
}

// Optional: plug in a callback for self-healing behavior (alerts, tickets, etc.)
log := common.CreateLoggingLLMWithCallback(
    "handler.go",
    "HandleRequest",
    func(analysis string) error {
        // e.g. send an email, create an issue, or store analysis in a DB
        return nil
    },
    "processing %s", r.URL.Path,
)
defer log.Print()
```

LLM analysis runs asynchronously in a background goroutine and never blocks the HTTP response path.

### Data Storage
```go
import "github.com/patdeg/common/datastore"

repo, err := datastore.NewRepository(ctx)
err = repo.Put(ctx, "User", key, user)
```

### Email Service
```go
import "github.com/patdeg/common/email"

service, err := email.NewService(config)
err = service.Send(ctx, message)
```

## 📚 Documentation

- [**Package Structure**](docs/PACKAGE_STRUCTURE.md) - Organization and architecture
- [**API Reference**](docs/API_REFERENCE.md) - Complete API documentation
- [**Logging Guide**](docs/LOGGING_GUIDE.md) - PII-safe logging helpers and patterns
- [**LLM Logging Reference**](docs/LOGGING_LLM.md) - `LoggingLLM` and automated error analysis
- [**Security Audit**](docs/SECURITY_AUDIT.md) - Security verification for public repo
- [**Examples**](examples/) - Working examples and tutorials

## 🔒 Security Notice

⚠️ **This is a PUBLIC repository**:
- ✅ No hardcoded credentials or secrets
- ✅ PII-safe logging by default  
- ✅ Input validation and sanitization
- ✅ Example data only (example.com domains)

## 🧪 Examples

See the [examples](examples/) directory:

```bash
# Basic usage with local implementations
cd examples/basic-usage && go run main.go

# Complete App Engine application
cd examples/appengine && go run main.go
```

## 🔧 Environment Configuration

The package automatically detects the environment and uses appropriate implementations:

**Development** (default): Local/in-memory services
**Production**: Cloud services (set `ENVIRONMENT=production`)

### Core Variables
```bash
# Authentication
GOOGLE_OAUTH_CLIENT_ID=xxxxx.apps.googleusercontent.com
GOOGLE_OAUTH_CLIENT_SECRET=your-secret
ADMIN_EMAILS=admin@example.com

# Google Cloud
PROJECT_ID=your-project-id
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json

# Email
SENDGRID_API_KEY=SG.example_fake_key.replace_with_real_key
FROM_EMAIL=noreply@example.com

# Optional: LLM-assisted error analysis
COMMON_LLM_API_KEY=gsk_example_fake_key_replace_with_real
COMMON_LLM_MODEL=meta-llama/llama-4-scout-17b-16e-instruct
COMMON_LLM_BASE_URL=https://api.groq.com/openai/v1
```

## 📈 Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                        │
├─────────────────────────────────────────────────────────────┤
│  Frontend │ API │ Search │ Monitor │ Import/Export          │
├─────────────────────────────────────────────────────────────┤
│              Business Logic Layer                            │
├─────────────────────────────────────────────────────────────┤
│     Tenant │ RBAC │ Payment │ Email │ Tasks                 │
├─────────────────────────────────────────────────────────────┤
│                Infrastructure Layer                          │
├─────────────────────────────────────────────────────────────┤
│   Datastore │ BigQuery │ Auth │ Cache │ Storage             │
├─────────────────────────────────────────────────────────────┤
│                   Core Utilities                             │
├─────────────────────────────────────────────────────────────┤
│  Logging │ Convert │ Crypto │ Files │ URLs │ Debug          │
└─────────────────────────────────────────────────────────────┘
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes (no sensitive data!)
4. Push to the branch
5. Open a Pull Request

## 📝 License

Apache License 2.0 - see [LICENSE](LICENSE) file for details.
