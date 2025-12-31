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

package common

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// AnalysisCallback is called when LLM analysis completes successfully.
// It receives the analysis result text and can perform custom actions like
// storing it in a database or sending notifications.
type AnalysisCallback func(analysis string) error

// Demeterics tag keys for analytics and tracking.
// See https://demeterics.ai/docs/prompt for full documentation.
const (
	// Business tags
	TagApp     = "APP"     // Application name
	TagFlow    = "FLOW"    // Flow/feature name (e.g., "checkout.payment")
	TagProduct = "PRODUCT" // Product identifier
	TagCompany = "COMPANY" // Company/tenant identifier
	TagUnit    = "UNIT"    // Business unit

	// User/Session tags
	TagUser    = "USER"    // User identifier (anonymized)
	TagSession = "SESSION" // Session identifier
	TagMarket  = "MARKET"  // Market/region

	// Technical tags
	TagVariant = "VARIANT" // A/B test variant
	TagVersion = "VERSION" // App version
	TagEnv     = "ENV"     // Environment (production, staging, dev)
	TagProject = "PROJECT" // GCP project or similar
)

// LoggingLLM captures standard log output alongside a markdown summary for a
// single logical operation. The summary is later available for printing or as
// additional context for LLM debugging.
type LoggingLLM struct {
	fileName string
	funcName string

	startTime time.Time

	mu               sync.Mutex
	summary          strings.Builder
	analysisOnce     sync.Once
	lastErrorText    string
	analysisCallback AnalysisCallback
	tags             map[string]string
}

var (
	llmHTTPClient = &http.Client{Timeout: 60 * time.Second}

	// Global throttle for LLM analysis to prevent duplicate feedback on the same error.
	// Key: hash of (fileName + funcName), Value: time of last analysis.
	analysisThrottleMu    sync.Mutex
	analysisThrottleCache = make(map[string]time.Time)

	// AnalysisThrottleDuration controls how long to suppress duplicate error analyses.
	// Default: 60 minutes. Set to 0 to disable throttling.
	AnalysisThrottleDuration = 60 * time.Minute
)

// CreateLoggingLLM constructs a LoggingLLM for the provided file and function
// names. The optional format string is logged immediately as an Info entry to
// seed the markdown summary.
func CreateLoggingLLM(fileName, funcName, format string, v ...interface{}) *LoggingLLM {
	return CreateLoggingLLMWithCallback(fileName, funcName, nil, format, v...)
}

// CreateLoggingLLMWithCallback constructs a LoggingLLM with a custom callback
// that will be invoked when LLM analysis completes successfully.
// The callback receives the analysis text and can store it, send notifications, etc.
func CreateLoggingLLMWithCallback(fileName, funcName string, callback AnalysisCallback, format string, v ...interface{}) *LoggingLLM {
	logger := &LoggingLLM{
		fileName:         fileName,
		funcName:         funcName,
		startTime:        time.Now(),
		analysisCallback: callback,
	}

	logger.mu.Lock()
	logger.summary.WriteString(fmt.Sprintf("### %s (%s)\n\n", funcName, fileName))
	logger.summary.WriteString(fmt.Sprintf("_Started %s_\n\n", logger.startTime.UTC().Format(time.RFC3339)))
	logger.mu.Unlock()

	if format != "" {
		logger.Info(format, v...)
	}

	return logger
}

// WithTags sets Demeterics metadata tags for analytics tracking.
// Tags are prepended to the LLM prompt as /// KEY value lines and stripped
// by Demeterics before forwarding to the provider (no token cost).
// Returns the logger for method chaining.
//
// Example:
//
//	log := common.CreateLoggingLLM("payment.go", "ProcessPayment", "starting").
//	    WithTags(map[string]string{
//	        common.TagApp:  "billing-service",
//	        common.TagFlow: "checkout.payment",
//	        common.TagEnv:  "production",
//	    })
func (l *LoggingLLM) WithTags(tags map[string]string) *LoggingLLM {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.tags == nil {
		l.tags = make(map[string]string)
	}
	for k, v := range tags {
		l.tags[k] = v
	}
	return l
}

// SetTag sets a single Demeterics metadata tag.
// Returns the logger for method chaining.
func (l *LoggingLLM) SetTag(key, value string) *LoggingLLM {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.tags == nil {
		l.tags = make(map[string]string)
	}
	l.tags[key] = value
	return l
}

// Debug logs a debug message and records it inside the markdown summary when
// ISDEBUG is enabled. The summary stores a PII-sanitized representation.
func (l *LoggingLLM) Debug(format string, v ...interface{}) {
	//Debug(format, v...)
	l.appendEntry("DEBUG", SanitizeMessage(fmt.Sprintf(format, v...)))
}

// DebugSafe logs a sanitized debug message and records it in the markdown
// summary when ISDEBUG is enabled.
func (l *LoggingLLM) DebugSafe(format string, v ...interface{}) {
	//DebugSafe(format, v...)
	l.appendEntry("DEBUG", SanitizeMessage(fmt.Sprintf(format, v...)))
}

// Info logs an informational message and records it in the markdown summary.
func (l *LoggingLLM) Info(format string, v ...interface{}) {
	Info(format, v...)
	l.appendEntry("INFO", fmt.Sprintf(format, v...))
}

// InfoSafe logs an informational message with PII sanitization applied and
// records the sanitized version in the markdown summary.
func (l *LoggingLLM) InfoSafe(format string, v ...interface{}) {
	InfoSafe(format, v...)
	l.appendEntry("INFO", SanitizeMessage(fmt.Sprintf(format, v...)))
}

// Warn logs a warning message and records it in the markdown summary.
func (l *LoggingLLM) Warn(format string, v ...interface{}) {
	Warn(format, v...)
	l.appendEntry("WARN", fmt.Sprintf(format, v...))
	l.Print()
}

// WarnSafe logs a warning message with sanitization applied and records the
// sanitized version in the markdown summary.
func (l *LoggingLLM) WarnSafe(format string, v ...interface{}) {
	WarnSafe(format, v...)
	l.appendEntry("WARN", SanitizeMessage(fmt.Sprintf(format, v...)))
	l.Print()
}

// ErrorNoAnalysis logs an error message and records it in the markdown summary
// WITHOUT triggering LLM analysis. Use this for errors that don't need AI debugging
// or when you want manual control over when analysis happens.
func (l *LoggingLLM) ErrorNoAnalysis(format string, v ...interface{}) {
	Error(format, v...)
	msg := fmt.Sprintf(format, v...)
	l.appendEntry("ERROR", msg)
	l.Print()
}

// ErrorNoAnalysisSafe logs an error with PII protection and records the sanitized
// message in the summary WITHOUT triggering LLM analysis.
func (l *LoggingLLM) ErrorNoAnalysisSafe(format string, v ...interface{}) {
	ErrorSafe(format, v...)
	msg := SanitizeMessage(fmt.Sprintf(format, v...))
	l.appendEntry("ERROR", msg)
	l.Print()
}

// Error logs an error message, records it in the markdown summary and
// triggers an asynchronous LLM analysis for additional guidance.
// If a callback was provided during creation, it will be called with the analysis result.
func (l *LoggingLLM) Error(format string, v ...interface{}) {
	Error(format, v...)
	msg := fmt.Sprintf(format, v...)
	l.appendEntry("ERROR", msg)
	l.triggerLLMAnalysis(msg)
}

// ErrorSafe logs an error with PII protection, records the sanitized message
// in the summary, and triggers the LLM analysis workflow.
// If a callback was provided during creation, it will be called with the analysis result.
func (l *LoggingLLM) ErrorSafe(format string, v ...interface{}) {
	ErrorSafe(format, v...)
	msg := SanitizeMessage(fmt.Sprintf(format, v...))
	l.appendEntry("ERROR", msg)
	l.triggerLLMAnalysis(msg)
}

// Print writes the current markdown summary to stdout.
func (l *LoggingLLM) Print() {
	Debug("=================================================================")
	fmt.Println(l.MarkdownSummary())
	Debug("=================================================================")
}

// MarkdownSummary returns the accumulated markdown summary including the
// elapsed duration since the logger was created.
func (l *LoggingLLM) MarkdownSummary() string {
	l.mu.Lock()
	defer l.mu.Unlock()

	duration := time.Since(l.startTime).Truncate(time.Millisecond)
	if duration < 0 {
		duration = 0
	}

	builder := strings.Builder{}
	builder.WriteString(l.summary.String())
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("_Duration %s_\n", duration))
	return builder.String()
}

// appendEntry stores a single markdown bullet with timestamp, level and
// message text.
func (l *LoggingLLM) appendEntry(level, message string) {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.summary.WriteString(fmt.Sprintf("- `%s` **%s** %s\n", time.Now().UTC().Format(time.RFC3339), level, trimmed))
}

// triggerLLMAnalysis runs error analysis asynchronously if an API key is
// configured. Multiple error calls will only trigger a single analysis run.
// Global throttling prevents duplicate analyses of the same error across instances.
func (l *LoggingLLM) triggerLLMAnalysis(message string) {
	if LLMAPIKey == "" {
		Debug("LLM_API_KEY not configured; skipping LLM analysis for %s.%s", l.fileName, l.funcName)
		return
	}

	// Check global throttle before proceeding
	throttleKey := computeThrottleKey(l.fileName, l.funcName, message)
	if isThrottled(throttleKey) {
		Debug("LLM analysis throttled for %s.%s (duplicate within %v)", l.fileName, l.funcName, AnalysisThrottleDuration)
		return
	}

	l.analysisOnce.Do(func() {
		l.lastErrorText = message
		// Mark as analyzed before running to prevent race conditions
		markAnalyzed(throttleKey)
		go l.runLLMAnalysis()
	})
}

// computeThrottleKey generates a hash key for throttling based on error context.
// Uses SHA-256 to create a consistent key from file and function only.
// This groups all errors from the same function together for throttling.
func computeThrottleKey(fileName, funcName, message string) string {
	combined := fmt.Sprintf("%s|%s", fileName, funcName)
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes (32 hex chars)
}

// isThrottled checks if an error with the given key was recently analyzed.
func isThrottled(key string) bool {
	if AnalysisThrottleDuration == 0 {
		return false // Throttling disabled
	}

	analysisThrottleMu.Lock()
	defer analysisThrottleMu.Unlock()

	lastTime, exists := analysisThrottleCache[key]
	if !exists {
		return false
	}

	// Check if still within throttle window
	if time.Since(lastTime) < AnalysisThrottleDuration {
		return true
	}

	// Expired, remove from cache
	delete(analysisThrottleCache, key)
	return false
}

// markAnalyzed records that an error was just analyzed.
func markAnalyzed(key string) {
	analysisThrottleMu.Lock()
	defer analysisThrottleMu.Unlock()

	analysisThrottleCache[key] = time.Now()

	// Cleanup old entries to prevent memory leak (keep max 1000 entries)
	if len(analysisThrottleCache) > 1000 {
		now := time.Now()
		for k, t := range analysisThrottleCache {
			if now.Sub(t) > AnalysisThrottleDuration {
				delete(analysisThrottleCache, k)
			}
		}
	}
}

func (l *LoggingLLM) runLLMAnalysis() {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	prompt := l.buildLLMPrompt()
	response, err := executeLLMRequest(ctx, prompt)
	if err != nil {
		l.Warn("LLM analysis failed for %s.%s: %v", l.fileName, l.funcName, err)
	        l.Print()
		return
	}

	l.appendEntry("LLM", response)
	l.Print()

	// Invoke callback if provided
	if l.analysisCallback != nil {
		if err := l.analysisCallback(response); err != nil {
			l.Warn("Analysis callback failed for %s.%s: %v", l.fileName, l.funcName, err)
		}
	}
}

// buildLLMPrompt assembles the system prompt, current summary, and source code
// snippet to provide rich context for the LLM.
func (l *LoggingLLM) buildLLMPrompt() string {
	var b strings.Builder

	// Prepend Demeterics tags (stripped before provider call, no token cost)
	l.mu.Lock()
	if len(l.tags) > 0 {
		// Write tags in a consistent order for readability
		tagOrder := []string{TagApp, TagFlow, TagProduct, TagCompany, TagUnit,
			TagUser, TagSession, TagMarket,
			TagVariant, TagVersion, TagEnv, TagProject}
		for _, key := range tagOrder {
			if val, ok := l.tags[key]; ok {
				b.WriteString(fmt.Sprintf("/// %s %s\n", key, val))
			}
		}
		// Write any custom tags not in the standard order
		for key, val := range l.tags {
			if !isStandardTag(key) {
				b.WriteString(fmt.Sprintf("/// %s %s\n", key, val))
			}
		}
		b.WriteString("\n")
	}
	l.mu.Unlock()

	b.WriteString("You are a senior Go engineer helping debug a failure.\n")
	b.WriteString("Provide probable root causes, code references, and actionable fixes.\n\n")
	b.WriteString(fmt.Sprintf("File: %s\nFunction: %s\n\n", l.fileName, l.funcName))
	if l.lastErrorText != "" {
		b.WriteString("### Latest error\n")
		b.WriteString(l.lastErrorText)
		b.WriteString("\n\n")
	}

	b.WriteString("### Recent log summary (markdown)\n")
	b.WriteString(l.MarkdownSummary())
	b.WriteString("\n")

	if snippet := l.loadSourceSnippet(); snippet != "" {
		b.WriteString("### Relevant source excerpt\n```go\n")
		b.WriteString(snippet)
		b.WriteString("\n```\n")
	}

	b.WriteString("\nFocus on the most likely fix and include any guardrails or tests to add.\n")
	return b.String()
}

func (l *LoggingLLM) loadSourceSnippet() string {
	if l.fileName == "" {
		return ""
	}

	data, err := os.ReadFile(l.fileName)
	if err != nil && !filepath.IsAbs(l.fileName) {
		data, err = os.ReadFile(filepath.Clean(l.fileName))
	}
	if err != nil {
		return ""
	}

	if len(data) > 20000 {
		data = data[:20000]
	}
	return string(data)
}

// executeLLMRequest sends the assembled prompt to the configured LLM provider.
func executeLLMRequest(ctx context.Context, prompt string) (string, error) {
	apiKey := LLMAPIKey
	if apiKey == "" {
		return "", fmt.Errorf("LLM_API_KEY not configured")
	}

	model := LLMModel
	if model == "" {
		model = "meta-llama/llama-4-scout-17b-16e-instruct"
	}

	baseURL := strings.TrimSuffix(LLMBaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.demeterics.com/groq/v1"
	}

	requestBody := struct {
		Model       string           `json:"model"`
		Messages    []llmChatMessage `json:"messages"`
		Temperature float64          `json:"temperature,omitempty"`
		MaxTokens   int              `json:"max_tokens,omitempty"`
	}{
		Model: model,
		Messages: []llmChatMessage{
			{Role: "system", Content: "You analyze Go backend errors and respond in markdown."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.2,
		MaxTokens:   2048,
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal LLM request: %w", err)
	}

	endpoint := baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create LLM request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := llmHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("LLM request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read LLM response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LLM provider returned %d: %s", resp.StatusCode, truncateForLog(string(body), 512))
	}

	var llmResp llmChatResponse
	if err := json.Unmarshal(body, &llmResp); err != nil {
		return "", fmt.Errorf("failed to parse LLM response: %w", err)
	}

	if len(llmResp.Choices) == 0 || llmResp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("LLM response missing content")
	}

	return llmResp.Choices[0].Message.Content, nil
}

// llmChatMessage mirrors OpenAI/Groq chat payload messages.
type llmChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// llmChatResponse covers the minimal fields used from the LLM provider.
type llmChatResponse struct {
	Choices []struct {
		Message llmChatMessage `json:"message"`
	} `json:"choices"`
}

func truncateForLog(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}

// isStandardTag checks if a tag key is one of the predefined Demeterics tags.
func isStandardTag(key string) bool {
	switch key {
	case TagApp, TagFlow, TagProduct, TagCompany, TagUnit,
		TagUser, TagSession, TagMarket,
		TagVariant, TagVersion, TagEnv, TagProject:
		return true
	}
	return false
}
