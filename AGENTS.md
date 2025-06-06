# Repo Overview

This repository contains small helper packages used in personal projects. Below is a high level description of each source file so that a GenAI agent can understand how they fit together.

## Root package (`common`)
- **common.go** – declares shared variables such as `ISDEBUG` and `VERSION`.
- **convert.go** – miscellaneous conversion utilities (string/number conversions, rounding, camel case, etc.).
- **convert_test.go** – tests for `CamelCase` from `convert.go`.
- **cookie.go** – visitor cookie helpers (create, read, clear) and `Visitor` struct.
- **cookie_test.go** – tests cookie creation and attributes.
- **crypt.go** – MD5 and CRC32 helpers plus AES based `Encrypt`/`Decrypt` functions.
- **debug.go** – utilities for dumping HTTP requests, responses and cookies for debugging.
- **file.go** – reads file contents.
- **interfaces.go** – generic routines to marshal/unmarshal JSON/XML and manage HTTP bodies.
- **logging.go** – simple Debug/Info/Error logging that respects `ISDEBUG`.
- **slice.go** – slice helper routines like `AddIfNotExists`.
- **url.go** – validation helper for HTTP/HTTPS URLs.
- **url_test.go** – tests for `url.go`.
- **web.go** – assorted web utilities: service account HTTP client, spam/bot detection, and helper HTML template rendering.

## `gcp` package
- **appengine.go** – extracts the App Engine version and sets the `VERSION` variable.
- **bigquery.go** – helpers to authenticate with BigQuery, create datasets/tables and stream rows.
- **datastore.go** – helper to fetch the first datastore entity in a query.
- **memcache.go** – wrappers for App Engine memcache operations on bytes and objects.
- **user.go** – datastore storage of users and roles.
- **user_test.go** – tests `EnsureUserExists` and `GetUserRole`.

## `ga` package
- **ga.go** – Google Analytics tracking helpers and structures for events.

These files depend on each other via the exported helpers. For example `web.go` uses `memcache`, `logging` and `convert`; BigQuery functions log via `logging.go` and use utilities from `crypt.go` and `convert.go`.

## `track` package
- **base.go** – package constants including dataset names and the 1×1 PNG used for pixel tracking.
- **config.go** – small helper `getEnv` used across the package.
- **types.go** – structures describing visits and robot hits.
- **adwords.go** – AdWords click tracking. Stores clicks in BigQuery and provides HTTP handlers.
- **bigquery_helpers.go** – helper `insertWithTableCreation` to stream data and create tables when needed.
- **bigquery_store.go** – streams `Visit` or `Event` data to BigQuery using helpers in `bigquery_helpers.go` and `common`.
- **bigquery_tables.go** – functions to create daily BigQuery tables for visits and events.
- **handlers.go** – HTTP handlers used to create the tables or record pixel/click tracking.
- **tracker.go** – core tracking logic for visits, events and bots. Uses memcache and the other helpers above.

## `auth` package
- **auth.go** – Google OAuth helper. Handles login redirect and callback, sets cookies and records login events via the `track` and `common` packages.

Overall, the `track` and `auth` packages build on the utilities from the root `common` package. BigQuery helpers and logging are shared across the repository.
