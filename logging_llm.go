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
}

var (
	llmHTTPClient = &http.Client{Timeout: 60 * time.Second}
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

// Debug logs a debug message and records it inside the markdown summary when
// ISDEBUG is enabled. The summary stores a PII-sanitized representation.
func (l *LoggingLLM) Debug(format string, v ...interface{}) {
	Debug(format, v...)
	if !ISDEBUG {
		return
	}
	l.appendEntry("DEBUG", SanitizeMessage(fmt.Sprintf(format, v...)))
}

// DebugSafe logs a sanitized debug message and records it in the markdown
// summary when ISDEBUG is enabled.
func (l *LoggingLLM) DebugSafe(format string, v ...interface{}) {
	DebugSafe(format, v...)
	if !ISDEBUG {
		return
	}
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
	Debug("-----------------------------------------------------------------")
	Warn(format, v...)
	Debug("-----------------------------------------------------------------")
	l.appendEntry("WARN", fmt.Sprintf(format, v...))
}

// WarnSafe logs a warning message with sanitization applied and records the
// sanitized version in the markdown summary.
func (l *LoggingLLM) WarnSafe(format string, v ...interface{}) {
	Debug("-----------------------------------------------------------------")
	WarnSafe(format, v...)
	Debug("-----------------------------------------------------------------")
	l.appendEntry("WARN", SanitizeMessage(fmt.Sprintf(format, v...)))
}

// ErrorNoAnalysis logs an error message and records it in the markdown summary
// WITHOUT triggering LLM analysis. Use this for errors that don't need AI debugging
// or when you want manual control over when analysis happens.
func (l *LoggingLLM) ErrorNoAnalysis(format string, v ...interface{}) {
	Debug("=================================================================")
	Error(format, v...)
	Debug("=================================================================")
	msg := fmt.Sprintf(format, v...)
	l.appendEntry("ERROR", msg)
}

// ErrorNoAnalysisSafe logs an error with PII protection and records the sanitized
// message in the summary WITHOUT triggering LLM analysis.
func (l *LoggingLLM) ErrorNoAnalysisSafe(format string, v ...interface{}) {
	Debug("=================================================================")
	ErrorSafe(format, v...)
	Debug("=================================================================")
	msg := SanitizeMessage(fmt.Sprintf(format, v...))
	l.appendEntry("ERROR", msg)
}

// Error logs an error message, records it in the markdown summary and
// triggers an asynchronous LLM analysis for additional guidance.
// If a callback was provided during creation, it will be called with the analysis result.
func (l *LoggingLLM) Error(format string, v ...interface{}) {
	Debug("=================================================================")
	Error(format, v...)
	Debug("=================================================================")
	msg := fmt.Sprintf(format, v...)
	l.appendEntry("ERROR", msg)
	l.triggerLLMAnalysis(msg)
}

// ErrorSafe logs an error with PII protection, records the sanitized message
// in the summary, and triggers the LLM analysis workflow.
// If a callback was provided during creation, it will be called with the analysis result.
func (l *LoggingLLM) ErrorSafe(format string, v ...interface{}) {
	Debug("=================================================================")
	ErrorSafe(format, v...)
	Debug("=================================================================")
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
func (l *LoggingLLM) triggerLLMAnalysis(message string) {
	if LLMAPIKey == "" {
		Debug("LLM_API_KEY not configured; skipping LLM analysis for %s.%s", l.fileName, l.funcName)
		return
	}

	l.analysisOnce.Do(func() {
		l.lastErrorText = message
		go l.runLLMAnalysis()
	})
}

func (l *LoggingLLM) runLLMAnalysis() {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	prompt := l.buildLLMPrompt()
	response, err := executeLLMRequest(ctx, prompt)
	if err != nil {
		l.Warn("LLM analysis failed for %s.%s: %v", l.fileName, l.funcName, err)
		return
	}

	l.appendEntry("LLM", response)

	Debug("=================================================================")
	ErrorSafe("ERROR Analysis: %v", response)
	Debug("=================================================================")

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
		baseURL = "https://api.groq.com/openai/v1"
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
