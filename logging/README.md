# Logging Package

The `logging` package provides advanced logging capabilities with automatic PII (Personally Identifiable Information) detection and sanitization.

## Features

- 🔒 **Automatic PII Detection**: Identifies and masks sensitive data patterns
- 📊 **Multiple Log Levels**: Debug, Info, Warn, Error, Fatal
- 🎯 **Structured Logging**: JSON output format support
- 🔧 **Configurable**: Environment variables and programmatic configuration
- 🚀 **High Performance**: Minimal overhead with cached regex patterns
- 🎨 **Customizable**: Add your own PII patterns

## Quick Example

```go
package main

import (
    "github.com/patdeg/common/logging"
)

func main() {
    // Create a logger
    logger := logging.NewLogger()
    
    // Log with automatic PII sanitization
    logger.Info("User logged in: john@example.com")
    // Output: [INFO] User logged in: j***n@***.com
    
    // Add custom pattern
    logger.Sanitizer.AddCustomPattern("account", `\bACCT\d{8}\b`)
    
    // Enable JSON output
    logger.SetJSONOutput(true)
}
```

## Files

- `safe.go` - Core logger implementation with PII-safe logging
- `sanitizer.go` - PII detection and masking engine
- `README.md` - This file

## See Also

- [Logging Guide](../docs/LOGGING_GUIDE.md) - Comprehensive usage guide
