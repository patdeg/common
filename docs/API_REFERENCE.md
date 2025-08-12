# API Reference

Complete API documentation for the common package.

## Table of Contents

- [Core Package](#core-package)
- [Logging Package](#logging-package)
- [Auth Package](#auth-package)
- [Datastore Package](#datastore-package)
- [BigQuery Package](#bigquery-package)
- [Email Package](#email-package)
- [Payment Package](#payment-package)
- [Search Package](#search-package)
- [Monitor Package](#monitor-package)

---

## Core Package

### Import
```go
import "github.com/patdeg/common"
```

### Constants
```go
const VERSION string  // Application version
```

### Functions

#### Logging Functions
```go
func Debug(format string, v ...interface{})
func Info(format string, v ...interface{})
func Warn(format string, v ...interface{})
func Error(format string, v ...interface{})

// PII-safe variants
func DebugSafe(format string, v ...interface{})
func InfoSafe(format string, v ...interface{})
func WarnSafe(format string, v ...interface{})
func ErrorSafe(format string, v ...interface{})
```

#### Type Conversion
```go
func StringToInt(s string) (int, error)
func IntToString(i int) string
func StringToBool(s string) (bool, error)
func BoolToString(b bool) string
func StringToFloat(s string) (float64, error)
func FloatToString(f float64) string
```

#### Slice Operations
```go
func Contains[T comparable](slice []T, item T) bool
func Unique[T comparable](slice []T) []T
func Filter[T any](slice []T, f func(T) bool) []T
func Map[T, U any](slice []T, f func(T) U) []U
```

#### Cryptography
```go
func Encrypt(key, plaintext string) (string, error)
func Decrypt(key, ciphertext string) (string, error)
func Hash(data string) string
func GenerateRandomKey() (string, error)
```

---

## Logging Package

### Import
```go
import "github.com/patdeg/common/logging"
```

### Types

#### LogSanitizer
```go
type LogSanitizer struct {
    // Configurable masking options
}

func NewLogSanitizer() *LogSanitizer
func (ls *LogSanitizer) Sanitize(message string) string
func (ls *LogSanitizer) AddCustomPattern(name, pattern string) error
```

### Functions
```go
func SetPIIProtection(enabled bool)
func AddCustomPIIPattern(name, pattern string) error
func SanitizeMessage(message string) string
```

---

## Auth Package

### Import
```go
import "github.com/patdeg/common/auth"
```

### Types

#### OAuth2Config
```go
type OAuth2Config struct {
    ClientID     string
    ClientSecret string
    RedirectURL  string
    Scopes       []string
}
```

#### User
```go
type User struct {
    ID       string `json:"id"`
    Email    string `json:"email"`
    Name     string `json:"name"`
    Picture  string `json:"picture"`
    IsAdmin  bool   `json:"is_admin"`
}
```

#### Session
```go
type Session struct {
    ID        string    `json:"id"`
    UserID    string    `json:"user_id"`
    ExpiresAt time.Time `json:"expires_at"`
}
```

### Functions
```go
func NewOAuth2Provider(provider string, config OAuth2Config) (*OAuth2Provider, error)
func (p *OAuth2Provider) GetAuthURL(state string) string
func (p *OAuth2Provider) Exchange(ctx context.Context, code string) (*User, error)
func CreateSession(user *User) (*Session, error)
func ValidateSession(sessionID string) (*User, error)
func GenerateCSRFToken() string
func ValidateCSRFToken(token string) bool
```

---

## Datastore Package

### Import
```go
import "github.com/patdeg/common/datastore"
```

### Interfaces

#### Repository
```go
type Repository interface {
    Get(ctx context.Context, kind, key string, dest interface{}) error
    Put(ctx context.Context, kind, key string, src interface{}) error
    Delete(ctx context.Context, kind, key string) error
    Query(ctx context.Context, query Query) ([]interface{}, error)
    Transaction(ctx context.Context, fn func(tx Transaction) error) error
}
```

#### Transaction
```go
type Transaction interface {
    Get(kind, key string, dest interface{}) error
    Put(kind, key string, src interface{}) error
    Delete(kind, key string) error
}
```

### Types

#### Query
```go
type Query struct {
    Kind    string
    Filters []Filter
    Orders  []Order
    Limit   int
    Offset  int
}
```

### Functions
```go
func NewRepository(ctx context.Context) (Repository, error)
func NewCloudRepository(ctx context.Context) (*CloudRepository, error)
func NewLocalRepository() *LocalRepository
```

---

## BigQuery Package

### Import
```go
import "github.com/patdeg/common/bigquery"
```

### Types

#### Config
```go
type Config struct {
    ProjectID     string
    DatasetID     string
    BatchSize     int
    BatchInterval time.Duration
}
```

#### Client
```go
type Client struct {
    // BigQuery client implementation
}
```

### Functions
```go
func NewClient(ctx context.Context, config Config) (*Client, error)
func (c *Client) InsertRow(ctx context.Context, tableID string, row interface{}, schema Schema) error
func (c *Client) InsertRowAsync(tableID string, row interface{})
func (c *Client) Query(ctx context.Context, sql string, params ...QueryParameter) (*RowIterator, error)
func (c *Client) Close(ctx context.Context) error
```

### Standard Schemas
```go
var StandardSchemas = struct {
    Telemetry Schema
    Audit     Schema
    Analytics Schema
}
```

---

## Email Package

### Import
```go
import "github.com/patdeg/common/email"
```

### Interfaces

#### Service
```go
type Service interface {
    Send(ctx context.Context, message *Message) error
    SendTemplate(ctx context.Context, template string, data interface{}, recipients []string) error
    SendBatch(ctx context.Context, messages []*Message) error
    ValidateEmail(email string) error
    GetProvider() string
}
```

### Types

#### Config
```go
type Config struct {
    Provider     string            // sendgrid, smtp, local
    APIKey       string
    FromEmail    string
    FromName     string
    SMTPHost     string
    SMTPPort     int
    Templates    map[string]string
}
```

#### Message
```go
type Message struct {
    From        Address
    To          []Address
    CC          []Address
    BCC         []Address
    ReplyTo     *Address
    Subject     string
    Text        string
    HTML        string
    Attachments []Attachment
    Headers     map[string]string
}
```

#### Address
```go
type Address struct {
    Email string `json:"email"`
    Name  string `json:"name,omitempty"`
}
```

### Functions
```go
func NewService(config Config) (Service, error)
func NewSendGridService(config Config) (*SendGridService, error)
func NewLocalService(config Config) *LocalService
```

---

## Payment Package

### Import
```go
import "github.com/patdeg/common/payment"
```

### Interfaces

#### Provider
```go
type Provider interface {
    CreateCustomer(ctx context.Context, customer *Customer) error
    GetCustomer(ctx context.Context, customerID string) (*Customer, error)
    CreateSubscription(ctx context.Context, sub *Subscription) error
    CancelSubscription(ctx context.Context, subscriptionID string, immediately bool) error
    ChargePayment(ctx context.Context, charge *Charge) error
    RefundPayment(ctx context.Context, refund *Refund) error
    HandleWebhook(ctx context.Context, payload []byte, signature string) (*WebhookEvent, error)
}
```

### Types

#### Customer
```go
type Customer struct {
    ID         string    `json:"id"`
    Email      string    `json:"email"`
    Name       string    `json:"name"`
    Currency   string    `json:"currency"`
    CreatedAt  time.Time `json:"created_at"`
}
```

#### Subscription
```go
type Subscription struct {
    ID         string             `json:"id"`
    CustomerID string             `json:"customer_id"`
    PlanID     string             `json:"plan_id"`
    Status     SubscriptionStatus `json:"status"`
    CreatedAt  time.Time         `json:"created_at"`
}
```

#### Plan
```go
type Plan struct {
    ID          string          `json:"id"`
    Name        string          `json:"name"`
    Amount      int64           `json:"amount"`
    Currency    string          `json:"currency"`
    Interval    BillingInterval `json:"interval"`
}
```

### Functions
```go
func NewManager(provider Provider) *Manager
func (m *Manager) CreateCustomer(ctx context.Context, email, name string) (*Customer, error)
func (m *Manager) Subscribe(ctx context.Context, customerID, planID string) (*Subscription, error)
func (m *Manager) CancelSubscription(ctx context.Context, subscriptionID string, immediately bool) error
```

---

## Search Package

### Import
```go
import "github.com/patdeg/common/search"
```

### Interfaces

#### Engine
```go
type Engine interface {
    Index(ctx context.Context, doc Document) error
    Search(ctx context.Context, query Query) (*Results, error)
    Delete(ctx context.Context, id string) error
    GetDocument(ctx context.Context, id string) (*Document, error)
}
```

### Types

#### Document
```go
type Document struct {
    ID        string                 `json:"id"`
    Index     string                 `json:"index"`
    Title     string                 `json:"title"`
    Content   string                 `json:"content"`
    Tags      []string              `json:"tags"`
    Metadata  map[string]interface{} `json:"metadata"`
    Timestamp time.Time             `json:"timestamp"`
}
```

#### Query
```go
type Query struct {
    Text      string      `json:"text"`
    Index     string      `json:"index"`
    Tags      []string    `json:"tags"`
    From      int         `json:"from"`
    Size      int         `json:"size"`
    Sort      []SortField `json:"sort"`
    Highlight bool        `json:"highlight"`
    Facets    []string    `json:"facets"`
}
```

#### Results
```go
type Results struct {
    Total  int                    `json:"total"`
    Hits   []Document            `json:"hits"`
    Facets map[string][]FacetItem `json:"facets"`
    Took   time.Duration         `json:"took"`
}
```

### Functions
```go
func NewInMemoryEngine() *InMemoryEngine
func NewQueryBuilder(text string) *QueryBuilder
```

---

## Monitor Package

### Import
```go
import "github.com/patdeg/common/monitor"
```

### Interfaces

#### HealthChecker
```go
type HealthChecker interface {
    Check(ctx context.Context) *HealthStatus
    Name() string
}
```

### Types

#### HealthStatus
```go
type HealthStatus struct {
    Status      Status                 `json:"status"`
    Message     string                 `json:"message"`
    Details     map[string]interface{} `json:"details"`
    LastChecked time.Time             `json:"last_checked"`
    Duration    time.Duration         `json:"duration_ms"`
}
```

#### HealthReport
```go
type HealthReport struct {
    Status    Status                   `json:"status"`
    Checks    map[string]*HealthStatus `json:"checks"`
    System    *SystemMetrics          `json:"system"`
    Timestamp time.Time               `json:"timestamp"`
}
```

### Functions
```go
func NewMonitor(checkPeriod time.Duration) *Monitor
func (m *Monitor) AddChecker(checker HealthChecker)
func (m *Monitor) GetHealth() *HealthReport
func (m *Monitor) ServeHTTP(w http.ResponseWriter, r *http.Request)
```

### Built-in Checkers
```go
func NewDatabaseChecker(name string, ping func(context.Context) error) *DatabaseChecker
func NewHTTPChecker(name, url string) *HTTPChecker
func NewDiskSpaceChecker(path string, threshold float64) *DiskSpaceChecker
```

---

## Error Handling

All packages follow consistent error handling:

```go
// Sentinel errors
var (
    ErrNotFound      = errors.New("not found")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrInvalidInput  = errors.New("invalid input")
    ErrTimeout       = errors.New("timeout")
)

// Error wrapping
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Error checking
if errors.Is(err, ErrNotFound) {
    // Handle not found
}
```

## Context Usage

All long-running operations accept a context:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := operation(ctx)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        // Handle timeout
    }
}
```

## Best Practices

1. **Always use contexts** for cancellation and timeouts
2. **Check errors** immediately after function calls
3. **Use interfaces** for flexibility and testing
4. **Validate inputs** before processing
5. **Log appropriately** using PII-safe functions
6. **Handle panics** in goroutines
7. **Close resources** using defer

---

*For more examples, see the [examples](../examples/) directory.*