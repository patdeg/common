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

// Package email provides a flexible email service abstraction that supports
// multiple providers (SendGrid, SMTP, local development).
package email

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/patdeg/common"
)

// Service defines the email service interface
type Service interface {
	// Send sends an email message
	Send(ctx context.Context, message *Message) error

	// SendTemplate sends a templated email
	SendTemplate(ctx context.Context, templateName string, data interface{}, recipients []string) error

	// SendBatch sends multiple emails in batch
	SendBatch(ctx context.Context, messages []*Message) error

	// ValidateEmail validates an email address
	ValidateEmail(email string) error

	// GetProvider returns the current provider name
	GetProvider() string
}

// Message represents an email message
type Message struct {
	From         Address                `json:"from"`
	To           []Address              `json:"to"`
	CC           []Address              `json:"cc,omitempty"`
	BCC          []Address              `json:"bcc,omitempty"`
	ReplyTo      *Address               `json:"reply_to,omitempty"`
	Subject      string                 `json:"subject"`
	Text         string                 `json:"text,omitempty"`
	HTML         string                 `json:"html,omitempty"`
	Attachments  []Attachment           `json:"attachments,omitempty"`
	Headers      map[string]string      `json:"headers,omitempty"`
	Metadata     map[string]string      `json:"metadata,omitempty"`
	TemplateID   string                 `json:"template_id,omitempty"`
	TemplateData map[string]interface{} `json:"template_data,omitempty"`
}

// Address represents an email address
type Address struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// Attachment represents an email attachment
type Attachment struct {
	Content     string `json:"content"`              // Base64 encoded content
	Type        string `json:"type"`                 // MIME type
	Filename    string `json:"filename"`             // File name
	Disposition string `json:"disposition"`          // inline or attachment
	ContentID   string `json:"content_id,omitempty"` // For inline attachments
}

// Config contains email service configuration
type Config struct {
	Provider     string            // sendgrid, smtp, local
	APIKey       string            // For API-based providers
	FromEmail    string            // Default from email
	FromName     string            // Default from name
	SMTPHost     string            // For SMTP provider
	SMTPPort     int               // For SMTP provider
	SMTPUser     string            // For SMTP provider
	SMTPPassword string            // For SMTP provider
	Templates    map[string]string // Template name -> template content
	IsDev        bool              // Development mode flag
}

// SendGridService implements Service using SendGrid
type SendGridService struct {
	config    Config
	apiKey    string
	fromEmail string
	fromName  string
	client    *http.Client
}

// LocalService implements Service for local development
type LocalService struct {
	config   Config
	messages []*Message // Store messages for inspection
}

// NewService creates a new email service based on configuration
func NewService(config Config) (Service, error) {
	// Auto-detect configuration from environment if not provided
	if config.Provider == "" {
		config.Provider = os.Getenv("EMAIL_PROVIDER")
		if config.Provider == "" {
			if os.Getenv("SENDGRID_API_KEY") != "" {
				config.Provider = "sendgrid"
			} else if os.Getenv("SMTP_HOST") != "" {
				config.Provider = "smtp"
			} else {
				config.Provider = "local"
			}
		}
	}

	if config.FromEmail == "" {
		config.FromEmail = os.Getenv("FROM_EMAIL")
		if config.FromEmail == "" {
			config.FromEmail = "noreply@example.com"
		}
	}

	if config.FromName == "" {
		config.FromName = os.Getenv("FROM_NAME")
		if config.FromName == "" {
			config.FromName = "Application"
		}
	}

	switch config.Provider {
	case "sendgrid":
		return NewSendGridService(config)
	case "smtp":
		// TODO: Implement SMTP service
		return NewLocalService(config), nil
	case "local":
		return NewLocalService(config), nil
	default:
		return nil, fmt.Errorf("unknown email provider: %s", config.Provider)
	}
}

// NewSendGridService creates a new SendGrid email service
func NewSendGridService(config Config) (*SendGridService, error) {
	if config.APIKey == "" {
		config.APIKey = os.Getenv("SENDGRID_API_KEY")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("SendGrid API key is required")
	}

	return &SendGridService{
		config:    config,
		apiKey:    config.APIKey,
		fromEmail: config.FromEmail,
		fromName:  config.FromName,
		client:    &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Send sends an email via SendGrid
func (s *SendGridService) Send(ctx context.Context, message *Message) error {
	// Set default from if not specified
	if message.From.Email == "" {
		message.From.Email = s.fromEmail
		message.From.Name = s.fromName
	}

	// Build SendGrid request
	sgReq := s.buildSendGridRequest(message)

	// Marshal to JSON
	data, err := json.Marshal(sgReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.sendgrid.com/v3/mail/send", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("SendGrid error (status %d): %v", resp.StatusCode, errResp)
	}

	common.Info("[EMAIL] Sent email via SendGrid: %s to %d recipients", message.Subject, len(message.To))
	return nil
}

// SendTemplate sends a templated email via SendGrid
func (s *SendGridService) SendTemplate(ctx context.Context, templateName string, data interface{}, recipients []string) error {
	// Build template content
	tmplContent, ok := s.config.Templates[templateName]
	if !ok {
		return fmt.Errorf("template not found: %s", templateName)
	}

	// Parse and execute template
	tmpl, err := template.New(templateName).Parse(tmplContent)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	var htmlBuf, textBuf bytes.Buffer
	if err := tmpl.Execute(&htmlBuf, data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	// Create message
	message := &Message{
		Subject: fmt.Sprintf("Message from %s", s.fromName),
		HTML:    htmlBuf.String(),
		Text:    textBuf.String(),
	}

	// Add recipients
	for _, email := range recipients {
		message.To = append(message.To, Address{Email: email})
	}

	return s.Send(ctx, message)
}

// SendBatch sends multiple emails in batch
func (s *SendGridService) SendBatch(ctx context.Context, messages []*Message) error {
	// SendGrid supports batch sending through personalizations
	// For simplicity, we'll send them individually
	for _, msg := range messages {
		if err := s.Send(ctx, msg); err != nil {
			common.Error("[EMAIL] Failed to send batch email: %v", err)
			// Continue with other emails
		}
	}
	return nil
}

// ValidateEmail validates an email address
func (s *SendGridService) ValidateEmail(email string) error {
	if !isValidEmail(email) {
		return fmt.Errorf("invalid email address: %s", email)
	}
	return nil
}

// GetProvider returns the provider name
func (s *SendGridService) GetProvider() string {
	return "sendgrid"
}

// buildSendGridRequest builds a SendGrid API request
func (s *SendGridService) buildSendGridRequest(message *Message) map[string]interface{} {
	// Build personalizations
	personalizations := []map[string]interface{}{{
		"to": convertAddresses(message.To),
	}}

	if len(message.CC) > 0 {
		personalizations[0]["cc"] = convertAddresses(message.CC)
	}

	if len(message.BCC) > 0 {
		personalizations[0]["bcc"] = convertAddresses(message.BCC)
	}

	// Build content
	var content []map[string]string
	if message.Text != "" {
		content = append(content, map[string]string{
			"type":  "text/plain",
			"value": message.Text,
		})
	}
	if message.HTML != "" {
		content = append(content, map[string]string{
			"type":  "text/html",
			"value": message.HTML,
		})
	}

	req := map[string]interface{}{
		"personalizations": personalizations,
		"from": map[string]string{
			"email": message.From.Email,
			"name":  message.From.Name,
		},
		"subject": message.Subject,
		"content": content,
	}

	// Add reply-to if specified
	if message.ReplyTo != nil {
		req["reply_to"] = map[string]string{
			"email": message.ReplyTo.Email,
			"name":  message.ReplyTo.Name,
		}
	}

	// Add attachments
	if len(message.Attachments) > 0 {
		var attachments []map[string]string
		for _, att := range message.Attachments {
			attachments = append(attachments, map[string]string{
				"content":     att.Content,
				"type":        att.Type,
				"filename":    att.Filename,
				"disposition": att.Disposition,
				"content_id":  att.ContentID,
			})
		}
		req["attachments"] = attachments
	}

	return req
}

// NewLocalService creates a new local email service for development
func NewLocalService(config Config) *LocalService {
	return &LocalService{
		config:   config,
		messages: make([]*Message, 0),
	}
}

// Send logs the email locally
func (s *LocalService) Send(ctx context.Context, message *Message) error {
	// Set default from if not specified
	if message.From.Email == "" {
		message.From.Email = s.config.FromEmail
		message.From.Name = s.config.FromName
	}

	// Store message
	s.messages = append(s.messages, message)

	// Log the email
	common.Info("[LOCAL_EMAIL] Email queued:")
	common.Info("  From: %s <%s>", message.From.Name, message.From.Email)
	common.Info("  To: %v", formatAddresses(message.To))
	if len(message.CC) > 0 {
		common.Info("  CC: %v", formatAddresses(message.CC))
	}
	if len(message.BCC) > 0 {
		common.Info("  BCC: %v", formatAddresses(message.BCC))
	}
	common.Info("  Subject: %s", message.Subject)

	if s.config.IsDev && message.HTML != "" {
		// In dev mode, save HTML to file for inspection
		filename := fmt.Sprintf("/tmp/email_%d.html", time.Now().Unix())
		os.WriteFile(filename, []byte(message.HTML), 0644)
		common.Info("  HTML saved to: %s", filename)
	}

	return nil
}

// SendTemplate sends a templated email locally
func (s *LocalService) SendTemplate(ctx context.Context, templateName string, data interface{}, recipients []string) error {
	message := &Message{
		Subject:      fmt.Sprintf("Template: %s", templateName),
		TemplateID:   templateName,
		TemplateData: map[string]interface{}{"data": data},
	}

	for _, email := range recipients {
		message.To = append(message.To, Address{Email: email})
	}

	return s.Send(ctx, message)
}

// SendBatch sends multiple emails locally
func (s *LocalService) SendBatch(ctx context.Context, messages []*Message) error {
	for _, msg := range messages {
		if err := s.Send(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

// ValidateEmail validates an email address
func (s *LocalService) ValidateEmail(email string) error {
	if !isValidEmail(email) {
		return fmt.Errorf("invalid email address: %s", email)
	}
	return nil
}

// GetProvider returns the provider name
func (s *LocalService) GetProvider() string {
	return "local"
}

// GetMessages returns all messages sent (for testing)
func (s *LocalService) GetMessages() []*Message {
	return s.messages
}

// ClearMessages clears all stored messages (for testing)
func (s *LocalService) ClearMessages() {
	s.messages = make([]*Message, 0)
}

// Helper functions

func isValidEmail(email string) bool {
	// Basic email validation
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return false
	}
	// Check for valid domain
	domainParts := strings.Split(parts[1], ".")
	if len(domainParts) < 2 {
		return false
	}
	return true
}

func convertAddresses(addresses []Address) []map[string]string {
	var result []map[string]string
	for _, addr := range addresses {
		m := map[string]string{"email": addr.Email}
		if addr.Name != "" {
			m["name"] = addr.Name
		}
		result = append(result, m)
	}
	return result
}

func formatAddresses(addresses []Address) string {
	var formatted []string
	for _, addr := range addresses {
		if addr.Name != "" {
			formatted = append(formatted, fmt.Sprintf("%s <%s>", addr.Name, addr.Email))
		} else {
			formatted = append(formatted, addr.Email)
		}
	}
	return strings.Join(formatted, ", ")
}

// AttachmentFromFile creates an attachment from a file
func AttachmentFromFile(filename string, contentType string) (*Attachment, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return &Attachment{
		Content:     base64.StdEncoding.EncodeToString(data),
		Type:        contentType,
		Filename:    filename,
		Disposition: "attachment",
	}, nil
}

// StandardTemplates provides standard email templates
var StandardTemplates = map[string]string{
	"welcome": `
<!DOCTYPE html>
<html>
<head>
    <title>Welcome</title>
</head>
<body>
    <h1>Welcome {{.Name}}!</h1>
    <p>Thank you for joining us.</p>
</body>
</html>`,
	"reset_password": `
<!DOCTYPE html>
<html>
<head>
    <title>Reset Password</title>
</head>
<body>
    <h1>Reset Your Password</h1>
    <p>Click <a href="{{.ResetLink}}">here</a> to reset your password.</p>
    <p>This link will expire in 1 hour.</p>
</body>
</html>`,
	"notification": `
<!DOCTYPE html>
<html>
<head>
    <title>Notification</title>
</head>
<body>
    <h1>{{.Title}}</h1>
    <p>{{.Message}}</p>
</body>
</html>`,
}
