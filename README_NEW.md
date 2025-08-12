# Common Go Package

[![Go Reference](https://pkg.go.dev/badge/github.com/patdeg/common.svg)](https://pkg.go.dev/github.com/patdeg/common)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

A comprehensive Go library providing reusable components for building scalable applications. This package offers production-ready utilities for authentication, data storage, payments, monitoring, and more.

## 🚀 Features

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

### Basic Logging with PII Protection

```go
import "github.com/patdeg/common"

func main() {
    // Regular logging
    common.Info("Application started")
    
    // PII-safe logging (automatically masks sensitive data)
    common.InfoSafe("User logged in: %s", userEmail)
    
    // Debug logging (only in development)
    common.Debug("Processing request: %v", requestData)
}
```

### Data Storage

```go
import "github.com/patdeg/common/datastore"

func main() {
    ctx := context.Background()
    
    // Initialize repository (auto-detects environment)
    repo, err := datastore.NewRepository(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    // Save entity
    user := &User{Email: "user@example.com", Name: "John Doe"}
    err = repo.Put(ctx, "User", user.Email, user)
    
    // Retrieve entity
    var retrieved User
    err = repo.Get(ctx, "User", "user@example.com", &retrieved)
}
```

### Email Service

```go
import "github.com/patdeg/common/email"

func main() {
    // Initialize email service
    service, err := email.NewService(email.Config{
        Provider: "sendgrid", // or "smtp", "local"
        // API key from environment variable
    })
    
    // Send email
    message := &email.Message{
        To:      []email.Address{{Email: "user@example.com"}},
        Subject: "Welcome!",
        HTML:    "<h1>Welcome to our service!</h1>",
    }
    
    err = service.Send(context.Background(), message)
}
```

### Health Monitoring

```go
import "github.com/patdeg/common/monitor"

func main() {
    // Create monitor
    mon := monitor.NewMonitor(30 * time.Second)
    
    // Add health checks
    mon.AddChecker(&monitor.PingChecker{})
    mon.AddChecker(monitor.NewDatabaseChecker("db", pingFunc))
    
    // Serve health endpoint
    http.Handle("/health", mon)
    http.ListenAndServe(":8080", nil)
}
```

## 📚 Package Documentation

### Core Utilities

| Package | Description | Import |
|---------|-------------|--------|
| `/` | Core utilities and helpers | `github.com/patdeg/common` |
| `/logging` | Enhanced logging with PII protection | `github.com/patdeg/common/logging` |
| `/auth` | Authentication and authorization | `github.com/patdeg/common/auth` |

### Data & Storage

| Package | Description | Import |
|---------|-------------|--------|
| `/datastore` | Generic data repository pattern | `github.com/patdeg/common/datastore` |
| `/bigquery` | BigQuery client with streaming | `github.com/patdeg/common/bigquery` |
| `/tasks` | Task queue abstraction | `github.com/patdeg/common/tasks` |

### Business Logic

| Package | Description | Import |
|---------|-------------|--------|
| `/tenant` | Multi-tenant support | `github.com/patdeg/common/tenant` |
| `/rbac` | Role-based access control | `github.com/patdeg/common/rbac` |
| `/payment` | Payment processing | `github.com/patdeg/common/payment` |

### Communications

| Package | Description | Import |
|---------|-------------|--------|
| `/email` | Email service abstraction | `github.com/patdeg/common/email` |

### Frontend & API

| Package | Description | Import |
|---------|-------------|--------|
| `/frontend` | Frontend asset management | `github.com/patdeg/common/frontend` |
| `/api` | HTTP client utilities | `github.com/patdeg/common/api` |
| `/search` | Full-text search engine | `github.com/patdeg/common/search` |

### Operations

| Package | Description | Import |
|---------|-------------|--------|
| `/monitor` | Health checks and metrics | `github.com/patdeg/common/monitor` |
| `/impexp` | Import/export utilities | `github.com/patdeg/common/impexp` |

### Platform Specific

| Package | Description | Import |
|---------|-------------|--------|
| `/gcp` | Google Cloud Platform utilities | `github.com/patdeg/common/gcp` |
| `/ga` | Google Analytics integration | `github.com/patdeg/common/ga` |

## 🏗️ Architecture

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

## 🔒 Security

This is a **PUBLIC** repository. We take security seriously:

- ✅ No hardcoded credentials or secrets
- ✅ PII-safe logging by default
- ✅ Input validation and sanitization
- ✅ SQL injection prevention
- ✅ XSS protection in web utilities
- ✅ CSRF token support

**Never commit**:
- API keys, tokens, or passwords
- Real email addresses or personal information
- Internal URLs or sensitive endpoints
- Customer data or business logic

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detection
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🔧 Configuration

### Environment Variables

The package uses environment variables for configuration. Set these before running your application:

#### Core Configuration
- `ENVIRONMENT`: Set to `production` for production mode (default: `development`)
- `DEBUG`: Enable debug logging (default: `false`)
- `LOG_PII_PROTECTION`: Disable PII protection in logs (default: `true`)

#### Google Cloud Platform
- `PROJECT_ID` or `GOOGLE_CLOUD_PROJECT`: GCP project ID
- `GOOGLE_APPLICATION_CREDENTIALS`: Path to service account JSON

#### Authentication
- `GOOGLE_OAUTH_CLIENT_ID`: OAuth client ID
- `GOOGLE_OAUTH_CLIENT_SECRET`: OAuth client secret
- `ADMIN_EMAILS`: Comma-separated list of admin emails

#### BigQuery
- `BQ_PROJECT_ID`: BigQuery project ID
- `BQ_DATASET`: Default dataset name

#### Email Service
- `EMAIL_PROVIDER`: Email provider (`sendgrid`, `smtp`, `local`)
- `SENDGRID_API_KEY`: SendGrid API key
- `FROM_EMAIL`: Default sender email
- `FROM_NAME`: Default sender name

#### Payment Processing
- `STRIPE_API_KEY`: Stripe API key
- `PADDLE_API_KEY`: Paddle API key

## 📋 Examples

See the [examples](examples/) directory for complete working examples:

```bash
# Run the App Engine example
cd examples/appengine
go run main.go

# Run the basic web server
cd examples/web-server
go run main.go
```

## 🚦 Environment Detection

The package automatically detects the environment:

- **Development**: Uses local/in-memory implementations
- **Production**: Uses cloud services (GCP, SendGrid, etc.)

Set `ENVIRONMENT=production` to force production mode.

## 📈 Versioning

We use [Semantic Versioning](https://semver.org/). For the versions available, see the [tags on this repository](https://github.com/patdeg/common/tags).

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please ensure:
- All tests pass
- No sensitive data is included
- Code follows Go best practices
- Documentation is updated

## 📝 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🗺️ Roadmap

- [ ] Add OpenTelemetry support
- [ ] GraphQL utilities
- [ ] WebSocket support
- [ ] Distributed tracing
- [ ] Circuit breaker patterns
- [ ] Feature flags system
- [ ] Workflow engine
- [ ] Message queue abstractions

## 📞 Support

- 🐛 Issues: [GitHub Issues](https://github.com/patdeg/common/issues)
- 📖 Docs: [pkg.go.dev](https://pkg.go.dev/github.com/patdeg/common)

---

Made with ❤️ by the team