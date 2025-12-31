# Logging LLM Reference

`logging_llm.go` extends the basic `common` logger with per-function markdown summaries and optional LLM-assisted debugging. Each `LoggingLLM` instance mirrors the standard logging helpers (`Debug`, `Info`, `Warn`, `Error`, and Safe variants) while recording structured notes that can be printed or piped into an LLM when errors occur.

## Configuration

Set the following environment variables (usually in your shell or deployment manifest) before launching your binary:

| Variable | Default | Description |
| --- | --- | --- |
| `COMMON_LLM_API_KEY` | _required_ | API key for Demeterics (`dmt_xxx` format). |
| `COMMON_LLM_MODEL` | `meta-llama/llama-4-scout-17b-16e-instruct` | Chat model used for automated analysis. |
| `COMMON_LLM_BASE_URL` | `https://api.demeterics.com/groq/v1` | Base URL for the OpenAI-compatible REST API (Demeterics proxy). |

If `COMMON_LLM_API_KEY` is not set, logging still works but no automated analysis will run when errors happen.

## Quick Start

```go
package handler

import (
	"net/http"

	"github.com/patdeg/common"
)

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	log := common.CreateLoggingLLM("handlers.go", "HandleRequest", "processing %s", r.URL.Path)
	defer log.Print() // emit the markdown summary at the end of the handler

	log.Debug("raw headers: %+v", r.Header)

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		log.WarnSafe("missing user id in query")
	}

	if err := process(userID); err != nil {
		log.ErrorSafe("failed processing user %s: %v", userID, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	log.Info("finished request")
}
```

The summary prints a concise markdown log of the run. If `log.Error*` is invoked and `COMMON_LLM_API_KEY` is configured, a background goroutine calls the configured LLM with the markdown summary and a snippet of the referenced file (`handlers.go`). The LLM response is appended to the summary under an `LLM` entry.

## API Reference

### Constructors

| Signature | Description |
| --- | --- |
| `func CreateLoggingLLM(fileName, funcName, format string, v ...interface{}) *LoggingLLM` | Creates a new structured logger. The initial `format` message seeds the summary with an Info entry (optional). |
| `func CreateLoggingLLMWithCallback(fileName, funcName string, callback AnalysisCallback, format string, v ...interface{}) *LoggingLLM` | Creates a logger with a custom callback that will be invoked when LLM analysis completes. |

### Callback Type

```go
type AnalysisCallback func(analysis string) error
```

Callbacks receive the LLM analysis text and can perform custom actions like:
- Store analysis in a database
- Create bug reports (GitHub Issues, Jira, etc.)
- Send notifications (Slack, email)
- Trigger alerts or workflows

Callbacks run asynchronously after LLM analysis completes. Errors in callbacks are logged but don't interrupt the main flow.

### Core Methods

All methods emit to stdout/stderr via the existing logging helpers and append markdown entries to the internal summary.

| Method | Behavior |
| --- | --- |
| `(*LoggingLLM) Debug(format string, v ...interface{})` | Logs only when `common.ISDEBUG` is true and writes a sanitized `DEBUG` entry. |
| `(*LoggingLLM) DebugSafe(format string, v ...interface{})` | Same as `Debug` but applies PII sanitization. |
| `(*LoggingLLM) Info(format string, v ...interface{})` | Always logs and records an `INFO` entry. |
| `(*LoggingLLM) InfoSafe(format string, v ...interface{})` | Sanitized variant of `Info`. |
| `(*LoggingLLM) Warn(format string, v ...interface{})` | Logs a warning and records a `WARN` entry. |
| `(*LoggingLLM) WarnSafe(format string, v ...interface{})` | Sanitized variant of `Warn`. |
| `(*LoggingLLM) Error(format string, v ...interface{})` | Logs an error, records an `ERROR` entry, and triggers LLM analysis (once). |
| `(*LoggingLLM) ErrorSafe(format string, v ...interface{})` | Sanitized variant of `Error`. |
| `(*LoggingLLM) Print()` | Writes the complete markdown summary (including duration) to stdout. |
| `(*LoggingLLM) MarkdownSummary() string` | Returns the accumulated markdown summary for custom routing. |

### Markdown Summary Format

- Header: `### functionName (fileName)` and start timestamp.
- Body: bullet list per log entry (`- 2025-02-19T08:32:10Z INFO finished request`).
- Footer: duration from creation to `Print`/`MarkdownSummary`.
- Optional: `LLM` bullet appended with the model’s response when automated analysis runs.

### LLM Analysis

1. Triggered automatically on the first `Error`/`ErrorSafe` call per logger instance.
2. Runs in a background goroutine with a 90s timeout.
3. Sends a prompt combining:
   - Latest error text.
   - Entire markdown summary to date.
   - Snippet of `fileName` (first ~20 KB).
4. Uses `COMMON_LLM_BASE_URL + /chat/completions` with the configured model.
5. Appends the model’s markdown response to the summary as an `LLM` entry.

Use the LLM output to surface suggested root causes or follow-up instrumentation automatically as part of error logs.
