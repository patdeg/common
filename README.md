# common

Shared helpers for Go projects.

Use the root package:

```go
import "github.com/patdeg/common"
```

Import subpackages, for example:

```go
import "github.com/patdeg/common/auth"
```

You can also import specialized helpers:
```go
import "github.com/patdeg/common/gcp"
import "github.com/patdeg/common/ga"
```

## Environment Variables

The packages read several configuration values from environment variables:

- `BQ_PROJECT_ID`: Google Cloud project used for storing visit and event data.
- `VISITS_DATASET`: BigQuery dataset name for visit tables.
- `EVENTS_DATASET`: BigQuery dataset name for event tables.
- `ADWORDS_PROJECT_ID`: Cloud project for AdWords click tracking.
- `ADWORDS_DATASET`: BigQuery dataset for AdWords click tables.
- `GOOGLE_OAUTH_CLIENT_ID`: OAuth client ID used by the auth package.
- `GOOGLE_OAUTH_CLIENT_SECRET`: OAuth client secret.
- `ADMIN_EMAILS`: comma-separated list of admin emails for login.

If a variable is unset, sensible defaults defined in the code will be used.

## App Engine Version

The `gcp` package includes a `Version` helper that fetches the App Engine
version ID from a request context. The function stores the major part of the
version string in the exported `common.VERSION` variable and returns it. Call
this helper during initialization so the application can log or act on the
deployed version via `common.VERSION`.

