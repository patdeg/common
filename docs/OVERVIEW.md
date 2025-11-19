# Common Package - Function Overview

This document provides a comprehensive overview of all functions in the `github.com/patdeg/common` package, organized by domain. Use this to identify helper functions you may want to import or replace with common's implementations.

**Last Updated:** 2025-11-18

---

## 🔧 Core Utilities

### Type Conversion (`convert.go`)

**Domain:** String/number conversions and formatting

- **`I2S(i int64) string`** - Converts int64 to decimal string representation
- **`B2S(b []byte) string`** - Converts null-terminated byte slice to string
- **`F2S(f float64) string`** - Converts float64 to fixed notation string (8 decimal places)
- **`S2F(s string) float64`** - Parses float from string, returns 0 on error
- **`S2I(s string) int64`** - Converts string to int64, returns 0 on parse failure
- **`S2B(s string) []byte`** - Converts string to byte slice
- **`ToString(x interface{}) string`** - Generic converter for int, int64, float64, string, or uses fmt.Sprint for others
- **`ToSQLString(x interface{}) string`** - Converts to string and escapes single quotes for SQL safety
- **`NULLIfEmpty(x string) string`** - Returns "NULL" literal if string is empty or "NaN"/"NANA"
- **`NumberToString(n int, sep rune) string`** - Formats integer with thousands separator
- **`ToNumber(s string) (bool, float64)`** - Attempts parsing string as float or int, returns success flag and value
- **`MonetaryToString(f float64) string`** - Formats float as currency with 2 decimal places
- **`TS(unixTime int64) string`** - Converts Unix millisecond timestamp to ANSI formatted time string
- **`Reverse(s string) string`** - Returns string with characters in reverse order
- **`Trunc500(s string) string`** - Truncates string to maximum 500 characters
- **`GetSuffix(s string, split string) string`** - Returns portion of string after final occurrence of split delimiter
- **`FirstPart(s string) string`** - Returns first semicolon-separated component
- **`CamelCase(txt string) string`** - Converts string with separators/punctuation to camel case
- **`Clean(txt string) string`** - Lowercases, replaces spaces with underscores, URL escapes result
- **`Round(num float64, precision int) float64`** - Rounds float to given precision using standard mathematical rounding

### Slice Operations (`slice.go`)

**Domain:** Working with string and generic slices

- **`AddIfNotExists(element string, list []string) []string`** - Appends element only if not already present, ensures uniqueness
- **`AddIfNotExistsGeneric(element interface{}, list []interface{}) []interface{}`** - Generic version for interface{} slices
- **`StringInSlice(a string, list []string) bool`** - Reports whether string exists in slice

### URL Utilities (`url.go`)

**Domain:** URL validation for security

- **`IsValidHTTPURL(dest string) bool`** - Verifies dest is absolute HTTP/HTTPS URL with valid scheme and host, prevents open redirects

### Input Validation (`validation/validation.go`)

**Domain:** Comprehensive input validation with security checks

**Validation Types:**
- **`ValidationError`** - Represents validation failure with field name, message, and error code
- **`ValidationErrors`** - Collection of multiple validation errors with combined error messages

**String Validators:**
- **`Required(field, value string) *ValidationError`** - Validates non-empty strings (trims whitespace)
- **`MaxLength(field, value string, max int) *ValidationError`** - Validates maximum string length
- **`MinLength(field, value string, min int) *ValidationError`** - Validates minimum string length
- **`Email(field, value string) *ValidationError`** - Validates RFC 5322 email addresses
- **`URL(field, value string) *ValidationError`** - Validates HTTP/HTTPS URLs
- **`UUID(field, value string) *ValidationError`** - Validates UUID v4 format
- **`ULID(field, value string) *ValidationError`** - Validates ULID format (github.com/oklog/ulid/v2)
- **`AlphanumericDashUnderscore(field, value string) *ValidationError`** - Validates [a-zA-Z0-9_-]+ pattern
- **`AlphanumericSpaceDashUnderscore(field, value string) *ValidationError`** - Validates [a-zA-Z0-9 _-]+ pattern
- **`ModelName(field, value string) *ValidationError`** - Validates AI model names (e.g., "gpt-4o", "llama-3.1-8b")
- **`CronExpression(field, value string) *ValidationError`** - Validates cron expressions (5-field format)
- **`OneOf(field, value string, allowed []string) *ValidationError`** - Validates value is in allowed list

**Numeric Validators:**
- **`IntRange(field string, value, min, max int) *ValidationError`** - Validates integer within range (inclusive)
- **`PositiveInt(field string, value int) *ValidationError`** - Validates value > 0
- **`NonNegativeInt(field string, value int) *ValidationError`** - Validates value >= 0

**Time Validators:**
- **`FutureTime(field string, value time.Time) *ValidationError`** - Validates time is in the future
- **`PastTime(field string, value time.Time) *ValidationError`** - Validates time is in the past

**Security Validators:**
- **`NoSQLInjection(field, value string) *ValidationError`** - Detects SQL injection patterns (defense-in-depth)
- **`NoXSS(field, value string) *ValidationError`** - Detects XSS patterns (defense-in-depth)

**Slice Validators:**
- **`MaxSliceLength[T any](field string, slice []T, max int) *ValidationError`** - Validates maximum slice length (generic)
- **`MinSliceLength[T any](field string, slice []T, min int) *ValidationError`** - Validates minimum slice length (generic)

**Validation Chaining:**
- **`NewValidator() *Validator`** - Creates validator for chaining multiple checks
- **`(*Validator) Add(err *ValidationError) *Validator`** - Adds validation error to chain
- **`(*Validator) Errors() error`** - Returns combined errors or nil
- **`(*Validator) HasErrors() bool`** - Checks if any errors exist

**Example Usage:**
```go
import "github.com/patdeg/common/validation"

v := validation.NewValidator()
v.Add(validation.Required("email", email))
v.Add(validation.Email("email", email))
v.Add(validation.MaxLength("email", email, 255))

if v.HasErrors() {
    return v.Errors() // Returns: "email: is required; email: must be a valid email address"
}
```

### File Operations (`file.go`)

**Domain:** Secure file reading with path validation

- **`ValidatePath(basePath, userPath string) (string, error)`** - Validates user path is safe and within basePath, prevents path traversal attacks including symlink attacks
- **`GetContent(c context.Context, filename string) (*[]byte, error)`** - Reads file and returns contents; requires ValidatePath() for user input

---

## 🔐 Security & Cryptography

### Cryptography (`crypt.go`)

**Domain:** Hashing, encryption, and secure ID generation

- **`SecureHash(data string) string`** - Generates SHA-256 hash for integrity checking (use for non-secret data)
- **`GenerateSecureID() (string, error)`** - Creates cryptographically secure 64-char hex random identifier (32 bytes/256 bits entropy)
- **`Hash(data string) uint32`** - Returns CRC32 checksum for non-security checksums
- **`Encrypt(c context.Context, key, message string) string`** - AES-256-GCM authenticated encryption, returns hex-encoded nonce+ciphertext
- **`Decrypt(c context.Context, key, message string) string`** - Decrypts AES-256-GCM message from Encrypt()
- **`MD5(data string) string`** - **DEPRECATED** MD5 hash (cryptographically broken, use SecureHash instead)

### CSRF Protection (`csrf/csrf.go`)

**Domain:** Cross-Site Request Forgery protection middleware

- **`NewTokenStore() *TokenStore`** - Creates new CSRF token store with automatic cleanup
- **`(*TokenStore) GenerateToken() (string, error)`** - Generates cryptographically secure 256-bit CSRF token
- **`(*TokenStore) ValidateToken(token string) bool`** - Validates token exists and hasn't expired (24h lifetime)
- **`(*TokenStore) Middleware(next http.Handler) http.Handler`** - HTTP middleware providing CSRF protection for state-changing methods
- **`(*TokenStore) StartCleanup(interval time.Duration)`** - Starts background goroutine to remove expired tokens

### Authentication (`auth/auth.go`)

**Domain:** OAuth2 and session management

- **`GetGoogleOAuthConfig(callbackURL string) *oauth2.Config`** - Returns configured Google OAuth2 config
- **`GetGitHubOAuthConfig(callbackURL string) *oauth2.Config`** - Returns configured GitHub OAuth2 config
- **`GetGoogleUserInfo(token string) (*GoogleUserInfo, error)`** - Fetches Google user profile using access token (secure: uses Authorization header)
- **`IsAdmin(email string) bool`** - Checks if email is in ADMIN_EMAILS environment variable

---

## 🧾 Logging & Monitoring

### Basic Logging (`logging.go`)

**Domain:** Standard log output with level prefixes

- **`Debug(format string, v ...interface{})`** - Writes debug message when ISDEBUG is true
- **`Info(format string, v ...interface{})`** - Writes informational message
- **`Warn(format string, v ...interface{})`** - Writes warning with "WARNING:" prefix
- **`Error(format string, v ...interface{})`** - Writes error with "ERROR:" prefix, optionally stores in Datastore if configured
- **`Fatal(format string, v ...interface{})`** - Logs fatal error and exits program with os.Exit(1)
- **`InitErrorDatastore() error`** - Initializes Datastore client for error logging when ERROR_DATASTORE_ENTITY is set

### PII-Safe Logging (`logging_enhanced.go`)

**Domain:** Logging with automatic PII sanitization

- **`DebugSafe(format string, v ...interface{})`** - Debug logging with PII sanitization applied
- **`InfoSafe(format string, v ...interface{})`** - Info logging with PII protection
- **`WarnSafe(format string, v ...interface{})`** - Warning logging with sanitization
- **`ErrorSafe(format string, v ...interface{})`** - Error logging with PII protection
- **`SanitizeMessage(message string) string`** - Applies PII sanitization to message
- **`AddCustomPIIPattern(name, pattern string) error`** - Adds custom regex pattern for PII detection
- **`SetPIIProtection(enabled bool)`** - Enables/disables global PII protection

### LLM-Assisted Logging (`logging_llm.go`)

**Domain:** Structured logging with AI-powered error analysis

- **`CreateLoggingLLM(fileName, funcName, format string, v ...interface{}) *LoggingLLM`** - Creates logger capturing markdown summary for operation
- **`CreateLoggingLLMWithCallback(fileName, funcName string, callback AnalysisCallback, format string, v ...interface{}) *LoggingLLM`** - Creates logger with custom callback for analysis results
- **`(*LoggingLLM) Debug(format string, v ...interface{})`** - Logs debug message to summary
- **`(*LoggingLLM) DebugSafe(format string, v ...interface{})`** - PII-safe debug logging
- **`(*LoggingLLM) Info(format string, v ...interface{})`** - Logs info message
- **`(*LoggingLLM) InfoSafe(format string, v ...interface{})`** - PII-safe info logging
- **`(*LoggingLLM) Warn(format string, v ...interface{})`** - Logs warning
- **`(*LoggingLLM) WarnSafe(format string, v ...interface{})`** - PII-safe warning
- **`(*LoggingLLM) Error(format string, v ...interface{})`** - Logs error and triggers async LLM analysis
- **`(*LoggingLLM) ErrorSafe(format string, v ...interface{})`** - PII-safe error with LLM analysis
- **`(*LoggingLLM) ErrorNoAnalysis(format string, v ...interface{})`** - Logs error WITHOUT LLM analysis
- **`(*LoggingLLM) ErrorNoAnalysisSafe(format string, v ...interface{})`** - PII-safe error without analysis
- **`(*LoggingLLM) Print()`** - Writes markdown summary to stdout
- **`(*LoggingLLM) MarkdownSummary() string`** - Returns accumulated markdown with duration

### PII Sanitization (`logging/sanitizer.go`)

**Domain:** Automatic PII detection and redaction

- **`NewLogSanitizer() *LogSanitizer`** - Creates sanitizer with built-in patterns for emails, IPs, credit cards, SSNs, phone numbers
- **`(*LogSanitizer) Sanitize(message string) string`** - Applies all patterns to redact PII from message
- **`(*LogSanitizer) AddCustomPattern(name, pattern string) error`** - Adds custom regex pattern for PII detection

### Health Monitoring (`monitor/health.go`)

**Domain:** HTTP health check endpoints

- **`HandleHealth(w http.ResponseWriter, r *http.Request)`** - Returns 200 OK with {"status":"healthy"} JSON response
- **`HandleReadiness(w http.ResponseWriter, r *http.Request)`** - Returns 200 OK with {"status":"ready"} JSON response for k8s readiness probes

---

## 🌐 Web & HTTP

### Cookie Management (`cookie.go`)

**Domain:** Secure cookie operations

- **`ClearCookie(w http.ResponseWriter, r *http.Request)`** - Removes visitor ID cookie by sending expired cookie
- **`DoesCookieExists(r *http.Request) bool`** - Checks if non-empty visitor ID cookie exists
- **`GetCookieID(w http.ResponseWriter, r *http.Request) string`** - Gets existing cookie or creates new secure cookie with HttpOnly, Secure (production), SameSite=Lax

### Web Utilities (`web.go`)

**Domain:** HTTP handlers, bot detection, security middleware

- **`HashIP(ip string) string`** - Hashes IP address with salt for GDPR/CCPA compliant logging (uses first 8 bytes of SHA-256)
- **`SecurityHeadersMiddleware(next http.Handler) http.Handler`** - Middleware setting security headers (X-Frame-Options, CSP, HSTS, X-Content-Type-Options, etc.)
- **`IsBot(r *http.Request) bool`** - Detects bots using user-agent parsing and custom bot list
- **`IsSpammer(r *http.Request) bool`** - Checks referrer against known spam domain blacklist
- **`IsHacker(r *http.Request) bool`** - Detects potential attacks (missing User-Agent, spam referrer, cached as hacker)
- **`GetServiceAccountClient(c context.Context) *http.Client`** - Creates HTTP client with App Engine service account credentials
- **`GetContentByUrl(c context.Context, url string) ([]byte, error)`** - Fetches URL content using service account client
- **`Message(w http.ResponseWriter, r *http.Request, message, redirect string, timeout int)`** - Renders HTML page with auto-redirect after timeout
- **`GetBodyResponse(r *http.Response) []byte`** - Reads response body and returns bytes

### Debug Utilities (`debug.go`)

**Domain:** HTTP debugging helpers (development only)

- **`DumpRequest(r *http.Request, withBody bool)`** - Logs complete HTTP request (including body if requested)
- **`DumpRequestOut(r *http.Request, withBody bool)`** - Logs outbound client request
- **`DumpResponse(c context.Context, r *http.Response)`** - Logs HTTP response while preserving body for reading
- **`DumpCookie(c context.Context, cookie *http.Cookie)`** - Logs cookie details
- **`DumpCookies(r *http.Request)`** - Logs all request cookies
- **`HandleEcho(w http.ResponseWriter, r *http.Request)`** - Debug endpoint echoing first 255 chars of POST body to logs

### HTTP Interfaces (`interfaces.go`)

**Domain:** JSON/XML response helpers

- **`WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) error`** - Writes JSON response with proper Content-Type header
- **`WriteXML(w http.ResponseWriter, statusCode int, data interface{}) error`** - Writes XML response with proper Content-Type header
- **`WriteError(w http.ResponseWriter, statusCode int, message string)`** - Writes JSON error response: {"error":"message"}

---

## 💾 Data Storage

### Datastore (`datastore/datastore.go`)

**Domain:** Google Cloud Datastore abstraction with local fallback

- **`NewRepository(ctx context.Context) (Repository, error)`** - Creates Datastore repository (cloud or in-memory based on environment)
- **`(*CloudRepository) Put(ctx context.Context, kind string, key interface{}, entity interface{}) error`** - Stores entity in Datastore
- **`(*CloudRepository) Get(ctx context.Context, kind string, key interface{}, entity interface{}) error`** - Retrieves entity from Datastore
- **`(*CloudRepository) Delete(ctx context.Context, kind string, key interface{}) error`** - Deletes entity from Datastore
- **`(*CloudRepository) Query(ctx context.Context, kind string, filters map[string]interface{}) ([]interface{}, error)`** - Queries entities with filters
- **`(*LocalRepository) Put/Get/Delete/Query(...)`** - In-memory implementations for development

### BigQuery (`bigquery/bigquery.go`)

**Domain:** Google BigQuery batch operations

- **`NewClient(ctx context.Context, projectID string) (*Client, error)`** - Creates BigQuery client
- **`(*Client) Insert(ctx context.Context, datasetID, tableID string, rows []interface{}) error`** - Inserts rows into BigQuery table
- **`(*Client) Query(ctx context.Context, query string) ([]map[string]interface{}, error)`** - Executes SQL query and returns results
- **`(*Client) CreateTable(ctx context.Context, datasetID, tableID string, schema interface{}) error`** - Creates table with schema
- **`(*Client) DeleteTable(ctx context.Context, datasetID, tableID string) error`** - Deletes table

### GCP Helpers (`gcp/`)

**Domain:** Google Cloud Platform integration utilities

#### App Engine (`gcp/appengine.go`)
- **`IsAppEngine() bool`** - Detects if running in App Engine environment
- **`GetProjectID() string`** - Returns project ID from GAE_APPLICATION or GOOGLE_CLOUD_PROJECT

#### Datastore (`gcp/datastore.go`)
- **`GetDatastoreKey(c context.Context, kind string, id int64) *datastore.Key`** - Creates Datastore key
- **`GetDatastoreEntity(c context.Context, key *datastore.Key, dst interface{}) error`** - Loads entity

#### BigQuery (`gcp/bigquery.go`)
- **`GetBigQueryClient(c context.Context) (*bigquery.Client, error)`** - Creates BigQuery client

#### Memcache (`gcp/memcache.go`)
- **`GetMemCacheString(c context.Context, key string) string`** - Gets string from memcache
- **`SetMemCacheString(c context.Context, key, value string, expiration int) error`** - Sets string in memcache with expiration

#### User Management (`gcp/user.go`)
- **`GetCurrentUser(c context.Context) (*User, error)`** - Gets current authenticated user
- **`IsAdmin(c context.Context, u *User) bool`** - Checks if user is admin

---

## 📧 Communication

### Email (`email/email.go`)

**Domain:** Email sending with SendGrid/SMTP

- **`NewService(config Config) (Service, error)`** - Creates email service (SendGrid, SMTP, or local mock)
- **`(*SendGridService) Send(ctx context.Context, msg Message) error`** - Sends email via SendGrid
- **`(*SMTPService) Send(ctx context.Context, msg Message) error`** - Sends email via SMTP
- **`(*LocalService) Send(ctx context.Context, msg Message) error`** - Mock email service logging to console

---

## 🔍 Search

### Search Engine (`search/search.go`)

**Domain:** In-memory full-text search with faceting

- **`NewEngine() *Engine`** - Creates new search engine
- **`(*Engine) Index(id string, doc Document) error`** - Indexes document for searching
- **`(*Engine) Search(query string, facets []string) (*SearchResult, error)`** - Searches indexed documents with optional faceting
- **`(*Engine) Delete(id string) error`** - Removes document from index
- **`(*Engine) Update(id string, doc Document) error`** - Updates indexed document

---

## 💰 Business Logic

### Multi-Tenancy (`tenant/tenant.go`)

**Domain:** Tenant isolation and limit management

- **`NewTenantManager(repo Repository) *Manager`** - Creates tenant manager
- **`(*Manager) GetTenant(ctx context.Context, id string) (*Tenant, error)`** - Retrieves tenant by ID
- **`(*Manager) CreateTenant(ctx context.Context, tenant *Tenant) error`** - Creates new tenant
- **`(*Manager) UpdateTenant(ctx context.Context, tenant *Tenant) error`** - Updates tenant
- **`(*Manager) CheckLimit(ctx context.Context, tenantID string, limitType string) (bool, error)`** - Checks if tenant is within limits

### RBAC (`rbac/rbac.go`)

**Domain:** Role-based access control

- **`NewRBACManager(repo Repository) *Manager`** - Creates RBAC manager
- **`(*Manager) AssignRole(ctx context.Context, userID, role string) error`** - Assigns role to user
- **`(*Manager) CheckPermission(ctx context.Context, userID, permission string) (bool, error)`** - Checks if user has permission
- **`(*Manager) GetUserRoles(ctx context.Context, userID string) ([]string, error)`** - Returns user's roles

### Payment Processing (`payment/payment.go`)

**Domain:** Subscription and payment management

- **`NewPaymentProvider(config Config) (Provider, error)`** - Creates payment provider (Stripe, Paddle, or mock)
- **`(*StripeProvider) CreateSubscription(ctx context.Context, sub Subscription) error`** - Creates subscription
- **`(*StripeProvider) CancelSubscription(ctx context.Context, subID string) error`** - Cancels subscription
- **`(*StripeProvider) HandleWebhook(ctx context.Context, payload []byte) (*Event, error)`** - Processes webhook events

---

## 📤 Data Import/Export

### Import/Export (`impexp/impexp.go`)

**Domain:** Data migration and backup/restore

- **`NewExporter(format string) (Exporter, error)`** - Creates exporter for JSON, CSV, or ZIP format
- **`(*JSONExporter) Export(ctx context.Context, data []interface{}) ([]byte, error)`** - Exports data as JSON
- **`(*CSVExporter) Export(ctx context.Context, data []interface{}) ([]byte, error)`** - Exports data as CSV
- **`NewImporter(format string) (Importer, error)`** - Creates importer for specified format
- **`(*JSONImporter) Import(ctx context.Context, data []byte) ([]interface{}, error)`** - Imports JSON data
- **`(*CSVImporter) Import(ctx context.Context, data []byte) ([]interface{}, error)`** - Imports CSV data

---

## ⚙️ Task Processing

### Tasks (`tasks/tasks.go`)

**Domain:** Background job processing with Cloud Tasks

- **`NewQueue(config Config) (Queue, error)`** - Creates task queue (cloud or local)
- **`(*CloudQueue) Enqueue(ctx context.Context, task Task) error`** - Adds task to Cloud Tasks queue
- **`(*CloudQueue) EnqueueBatch(ctx context.Context, tasks []Task) error`** - Adds multiple tasks efficiently
- **`(*LocalQueue) Enqueue(ctx context.Context, task Task) error`** - Processes task immediately in local mode

---

## 📊 Analytics & Tracking

### Google Analytics (`ga/ga.go`)

**Domain:** Google Analytics event tracking

- **`TrackEvent(ctx context.Context, r *http.Request, category, action, label string, value int) error`** - Sends event to Google Analytics
- **`TrackPageView(ctx context.Context, r *http.Request, path string) error`** - Tracks page view

### Custom Tracking (`track/`)

**Domain:** Internal event tracking and BigQuery storage

#### Tracker (`track/tracker.go`)
- **`NewTracker(config Config) *Tracker`** - Creates event tracker
- **`(*Tracker) TrackEvent(ctx context.Context, event Event) error`** - Records custom event

#### BigQuery Storage (`track/bigquery_store.go`)
- **`NewBigQueryStore(client *bigquery.Client, datasetID string) *Store`** - Creates BigQuery event store
- **`(*Store) StoreEvent(ctx context.Context, event Event) error`** - Stores event in BigQuery
- **`(*Store) StoreEvents(ctx context.Context, events []Event) error`** - Batch stores events

#### AdWords (`track/adwords.go`)
- **`TrackConversion(ctx context.Context, conversion Conversion) error`** - Tracks AdWords conversion

---

## 🎨 Frontend Support

### Asset Management (`frontend/assets.go`)

**Domain:** Static asset serving with versioning and caching

- **`NewAssetManager(basePath string) *AssetManager`** - Creates asset manager for serving static files
- **`(*AssetManager) ServeHTTP(w http.ResponseWriter, r *http.Request)`** - HTTP handler serving assets with proper headers
- **`(*AssetManager) GetAssetURL(path string) string`** - Returns versioned asset URL with SHA-256 hash
- **`(*AssetManager) LoadAssets() error`** - Pre-loads and hashes assets for production

---

## 🔌 API Client

### HTTP Client (`api/client.go`)

**Domain:** HTTP client with retry and rate limiting

- **`NewClient(config Config) *Client`** - Creates HTTP client with retry logic
- **`(*Client) Get(ctx context.Context, url string) (*Response, error)`** - GET request with retry
- **`(*Client) Post(ctx context.Context, url string, body interface{}) (*Response, error)`** - POST request with JSON body
- **`(*Client) Put(ctx context.Context, url string, body interface{}) (*Response, error)`** - PUT request with JSON body
- **`(*Client) Delete(ctx context.Context, url string) (*Response, error)`** - DELETE request with retry
- **`(*Client) Do(ctx context.Context, req *http.Request) (*Response, error)`** - Executes request with configured retry/timeout

---

## 📋 Environment Variables

### Configuration (`common.go`)

**Domain:** Global configuration from environment

- **`ISDEBUG bool`** - Debug mode flag from COMMON_DEBUG env var
- **`VERSION string`** - Application version (set via ldflags)
- **`LLMAPIKey string`** - API key for LLM provider from COMMON_LLM_API_KEY
- **`LLMModel string`** - LLM model identifier from COMMON_LLM_MODEL (default: llama-4-scout)
- **`LLMBaseURL string`** - LLM API endpoint from COMMON_LLM_BASE_URL (default: Groq)

---

## 🔒 Security Features

### Recent Security Enhancements

All security vulnerabilities identified in the 2025-11-18 audit have been remediated:

1. **CSRF Protection** - Complete middleware package with 79.4% test coverage
2. **Path Traversal Prevention** - ValidatePath() with symlink attack detection
3. **Secure Hashing** - SecureHash() and GenerateSecureID() replacing MD5
4. **PII Protection** - HashIP() for GDPR/CCPA compliant logging
5. **Security Headers** - SecurityHeadersMiddleware() for XSS/clickjacking protection
6. **OAuth Security** - Access tokens via Authorization header (not URL)

### Security Scanner

Run `./scripts/run-security-scans.sh` to execute:
- gosec (SAST)
- staticcheck
- govulncheck
- Secret scanning
- Test coverage analysis

---

## 📖 Usage Recommendations

### When to Import from Common

**Import when you need:**
- ✅ Type conversion utilities (convert.go)
- ✅ PII-safe logging (logging_enhanced.go)
- ✅ LLM-assisted error analysis (logging_llm.go)
- ✅ CSRF protection middleware (csrf/)
- ✅ Secure cookie management (cookie.go)
- ✅ Path traversal protection (file.go ValidatePath)
- ✅ HTTP security middleware (web.go SecurityHeadersMiddleware)
- ✅ Bot/spam detection (web.go)
- ✅ Datastore abstraction (datastore/)
- ✅ BigQuery batch operations (bigquery/)

### When to Keep Local Implementation

**Keep local if:**
- ❌ Your implementation has domain-specific customization
- ❌ You need features not provided by common
- ❌ Your code has different error handling requirements
- ❌ Performance profiling shows common's version is slower for your use case

### Migration Checklist

When replacing local helpers with common's versions:

1. Compare function signatures - ensure parameters match
2. Check error handling - common may return errors where local code doesn't
3. Verify environment variables - common uses specific env var names
4. Test PII handling - common's safe logging may sanitize differently
5. Review dependencies - ensure common's dependencies are acceptable
6. Run security scans - validate no new vulnerabilities introduced

---

## 📞 Support

- **Documentation:** See `docs/` directory for detailed guides
- **Security Issues:** Run `./scripts/run-security-scans.sh`
- **Examples:** Check `examples/` for working code samples
- **API Reference:** See `docs/API_REFERENCE.md`

**Package Version:** See VERSION in common.go (set at build time)
