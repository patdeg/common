# Basic Usage Example

This example demonstrates the fundamental features of the common package:

- Logging with PII protection
- Type conversion utilities  
- Slice operations
- Data storage (local in development)
- Email service (local mode)

## Running the Example

```bash
cd examples/basic-usage
go run main.go
```

## What This Example Shows

### 1. Logging with PII Protection
```go
// Regular logging
common.Info("Application started")

// PII-safe logging (automatically masks emails, etc.)
common.InfoSafe("User logged in: %s", userEmail)

// Debug logging (development only)
common.Debug("Debug information")
```

### 2. Type Conversions
```go
// Safe type conversions with error handling
num, err := common.StringToInt("123")
str := common.IntToString(456)
flag, err := common.StringToBool("true")
```

### 3. Slice Operations
```go
// Generic slice utilities
contains := common.Contains(slice, item)
unique := common.Unique(slice)
filtered := common.Filter(slice, predicate)
transformed := common.Map(slice, transformer)
```

### 4. Data Storage
```go
// Environment-aware repository
repo, err := datastore.NewRepository(ctx)
err = repo.Put(ctx, "User", key, user)
err = repo.Get(ctx, "User", key, &user)
```

### 5. Email Service
```go
// Multi-provider email service
service, err := email.NewService(config)
err = service.Send(ctx, message)
```

## Expected Output

The example will output information about each operation:

```
=== Common Package Basic Usage Example ===

1. Logging Examples:
   ✓ Logging examples completed

2. Type Conversion Examples:
   String '123' converted to int: 123
   Int 456 converted to string: '456'
   String 'true' converted to bool: true
   ✓ Type conversion examples completed

3. Slice Operations Examples:
   Slice contains 3: true
   Original: [1 2 3 2 4 3 5]
   Unique:   [1 2 3 4 5]
   Even numbers: [2 2 4]
   Doubled: [2 4 6 4 8 6 10]
   ✓ Slice operations examples completed

4. Data Storage Examples:
   ✓ Saved user: john.doe@example.com
   ✓ Retrieved user: John Doe (age: 30)
   ✓ Deleted user: john.doe@example.com
   ✓ Data storage examples completed

5. Email Service Examples:
   ✓ Email sent to: user@example.com
   ✓ Email service examples completed

=== Example completed successfully! ===
```

## Environment

This example works entirely with local/in-memory implementations and requires no external configuration. It's designed to run immediately without any setup.