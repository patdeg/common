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

## Running the Example

Set the required environment variables before starting the sample server:

```bash
export BQ_PROJECT_ID=your-gcp-project
export VISITS_DATASET=visits
export EVENTS_DATASET=events
export ADWORDS_PROJECT_ID=your-gcp-project
export ADWORDS_DATASET=adwords
export GOOGLE_OAUTH_CLIENT_ID=xxxxx.apps.googleusercontent.com
export GOOGLE_OAUTH_CLIENT_SECRET=your-secret
export ADMIN_EMAILS=user@example.com
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service_account.json
```

Run the App Engine example with:

```bash
go run ./examples/appengine
```

The server listens on the port specified by the `PORT` variable (defaults to
`8080`).

## BigQuery and OAuth Configuration

1. Create a Google Cloud project and enable the BigQuery API.
2. Create datasets named `VISITS_DATASET`, `EVENTS_DATASET` and
   `ADWORDS_DATASET` in the project identified by `BQ_PROJECT_ID`.
3. Download a service account key and set `GOOGLE_APPLICATION_CREDENTIALS` to
   the JSON file path.
4. In the Cloud Console, create OAuth client credentials for a web application
   and set `GOOGLE_OAUTH_CLIENT_ID` and `GOOGLE_OAUTH_CLIENT_SECRET` to the
   generated values.

## Security Updates

The project now validates redirect targets and uses AES‑GCM encryption for
sensitive data.

