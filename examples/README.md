# Examples

This directory contains working examples demonstrating how to use the common package components.

## Available Examples

- [**basic-usage**](basic-usage/) - Simple demonstration of core utilities
- [**appengine**](appengine/) - Complete App Engine application
- [**web-server**](web-server/) - Standalone web server with authentication
- [**data-processing**](data-processing/) - BigQuery and data storage examples
- [**payment-integration**](payment-integration/) - Payment processing with multiple providers
- [**monitoring**](monitoring/) - Health checks and metrics collection

## Running Examples

Each example directory contains:
- `main.go` - The example code
- `README.md` - Specific instructions and setup
- `.env.example` - Required environment variables

### Quick Start

1. Copy environment file:
```bash
cd examples/[example-name]
cp .env.example .env
# Edit .env with your configuration
```

2. Run the example:
```bash
go run main.go
```

### General Environment Variables

Most examples require these basic variables:

```bash
# Google Cloud (if using GCP features)
PROJECT_ID=your-project-id
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json

# Authentication (if using auth features)
GOOGLE_OAUTH_CLIENT_ID=xxxxx.apps.googleusercontent.com
GOOGLE_OAUTH_CLIENT_SECRET=your-secret
ADMIN_EMAILS=admin@example.com

# Email (if using email features)
EMAIL_PROVIDER=local
SENDGRID_API_KEY=your-key
FROM_EMAIL=noreply@example.com
FROM_NAME=Example App

# Environment
ENVIRONMENT=development
DEBUG=true
```

## Contributing Examples

When adding new examples:

1. Create a new directory with a descriptive name
2. Include `main.go`, `README.md`, and `.env.example`
3. Use only example.com domains and dummy data
4. Document all required environment variables
5. Test in both development and production modes
6. Keep examples focused and simple

## Security Notice

⚠️ **IMPORTANT**: This is a PUBLIC repository. Examples must:

- Use only example.com email addresses
- Never include real API keys or credentials
- Use dummy/placeholder data only
- Include .env in .gitignore
- Document security considerations

All examples are designed to work with mock data and can be safely shared publicly.