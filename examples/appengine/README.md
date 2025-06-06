This example demonstrates using the `auth` subpackage to require Google OAuth
login before serving a simple page.

Handlers
--------

- `/` handled by `HelloHandler` which greets authenticated users.
- `/goog_login` begins the OAuth flow using `auth.GoogleLoginHandler`.
- `/goog_callback` finishes the OAuth flow via `auth.GoogleCallbackHandler`.

Environment variables
---------------------

- `GOOGLE_OAUTH_CLIENT_ID` and `GOOGLE_OAUTH_CLIENT_SECRET` – credentials for a
  Google OAuth client.
- `ADMIN_EMAILS` – optional comma-separated list of admin emails.
- `PORT` – HTTP port to listen on (default `8080`).

Run `go run main.go` from this directory after setting the variables above.
