// Copyright 2025 Patrick Deglon
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package logging provides PII-safe logging functionality for applications.
// It includes automatic sanitization of personally identifiable information (PII)
// and structured logging capabilities.
package logging

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	// DebugLevel logs are typically only enabled in development
	DebugLevel LogLevel = iota
	// InfoLevel logs are informational messages
	InfoLevel
	// WarnLevel logs are warning messages
	WarnLevel
	// ErrorLevel logs are error messages
	ErrorLevel
	// FatalLevel logs are fatal messages that will cause the program to exit
	FatalLevel
)

// Logger provides PII-safe logging functionality
type Logger struct {
	mu            sync.RWMutex
	level         LogLevel
	sanitizer     *LogSanitizer
	isDebug       bool
	jsonOutput    bool
	includeSource bool
	prefix        string
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Source    string                 `json:"source,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

var (
	// DefaultLogger is the default logger instance
	DefaultLogger *Logger
	once          sync.Once
)

// init initializes the default logger
func init() {
	once.Do(func() {
		DefaultLogger = NewLogger()
	})
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	isDebug := os.Getenv("DEBUG") == "true" || os.Getenv("ISDEBUG") == "true"
	
	return &Logger{
		level:         InfoLevel,
		sanitizer:     NewLogSanitizer(),
		isDebug:       isDebug,
		jsonOutput:    os.Getenv("LOG_FORMAT") == "json",
		includeSource: os.Getenv("LOG_SOURCE") == "true",
		prefix:        "",
	}
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetJSONOutput enables or disables JSON output format
func (l *Logger) SetJSONOutput(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.jsonOutput = enabled
}

// SetPrefix sets a prefix for all log messages
func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}

// Debug logs a debug message with PII sanitization
func (l *Logger) Debug(format string, v ...interface{}) {
	if !l.isDebug {
		return
	}
	l.log(DebugLevel, format, v...)
}

// Info logs an informational message with PII sanitization
func (l *Logger) Info(format string, v ...interface{}) {
	l.log(InfoLevel, format, v...)
}

// Warn logs a warning message with PII sanitization
func (l *Logger) Warn(format string, v ...interface{}) {
	l.log(WarnLevel, format, v...)
}

// Error logs an error message with PII sanitization
func (l *Logger) Error(format string, v ...interface{}) {
	l.log(ErrorLevel, format, v...)
}

// Fatal logs a fatal message and exits the program
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.log(FatalLevel, format, v...)
	os.Exit(1)
}

// log handles the actual logging with PII sanitization
func (l *Logger) log(level LogLevel, format string, v ...interface{}) {
	l.mu.RLock()
	if level < l.level {
		l.mu.RUnlock()
		return
	}
	
	jsonOutput := l.jsonOutput
	includeSource := l.includeSource
	prefix := l.prefix
	l.mu.RUnlock()

	// Format the message
	message := fmt.Sprintf(format, v...)
	
	// Sanitize the message to remove PII
	message = l.sanitizer.Sanitize(message)
	
	// Add prefix if set
	if prefix != "" {
		message = prefix + " " + message
	}
	
	// Get source information if enabled
	var source string
	if includeSource {
		_, file, line, ok := runtime.Caller(2)
		if ok {
			// Simplify the file path
			parts := strings.Split(file, "/")
			if len(parts) > 2 {
				file = strings.Join(parts[len(parts)-2:], "/")
			}
			source = fmt.Sprintf("%s:%d", file, line)
		}
	}
	
	// Output the log
	if jsonOutput {
		entry := LogEntry{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Level:     levelToString(level),
			Message:   message,
			Source:    source,
		}
		
		data, _ := json.Marshal(entry)
		log.Println(string(data))
	} else {
		levelStr := levelToString(level)
		if source != "" {
			log.Printf("[%s] %s (%s)\n", levelStr, message, source)
		} else {
			log.Printf("[%s] %s\n", levelStr, message)
		}
	}
}

// levelToString converts a LogLevel to its string representation
func levelToString(level LogLevel) string {
	switch level {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Package-level convenience functions that use the default logger

// Debug logs a debug message with PII sanitization
func Debug(format string, v ...interface{}) {
	DefaultLogger.Debug(format, v...)
}

// Info logs an informational message with PII sanitization
func Info(format string, v ...interface{}) {
	DefaultLogger.Info(format, v...)
}

// Warn logs a warning message with PII sanitization
func Warn(format string, v ...interface{}) {
	DefaultLogger.Warn(format, v...)
}

// Error logs an error message with PII sanitization
func Error(format string, v ...interface{}) {
	DefaultLogger.Error(format, v...)
}

// Fatal logs a fatal message and exits the program
func Fatal(format string, v ...interface{}) {
	DefaultLogger.Fatal(format, v...)
}

// SetLevel sets the minimum log level for the default logger
func SetLevel(level LogLevel) {
	DefaultLogger.SetLevel(level)
}

// SetJSONOutput enables or disables JSON output for the default logger
func SetJSONOutput(enabled bool) {
	DefaultLogger.SetJSONOutput(enabled)
}

// SetPrefix sets a prefix for all log messages from the default logger
func SetPrefix(prefix string) {
	DefaultLogger.SetPrefix(prefix)
}