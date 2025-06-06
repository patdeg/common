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
