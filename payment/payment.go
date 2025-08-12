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

// Package payment provides subscription and payment processing with
// support for multiple payment providers.
package payment

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/patdeg/common"
)

// Provider defines the payment provider interface
type Provider interface {
	// CreateCustomer creates a new customer
	CreateCustomer(ctx context.Context, customer *Customer) error
	
	// GetCustomer retrieves customer details
	GetCustomer(ctx context.Context, customerID string) (*Customer, error)
	
	// UpdateCustomer updates customer information
	UpdateCustomer(ctx context.Context, customer *Customer) error
	
	// CreateSubscription creates a new subscription
	CreateSubscription(ctx context.Context, sub *Subscription) error
	
	// GetSubscription retrieves subscription details
	GetSubscription(ctx context.Context, subscriptionID string) (*Subscription, error)
	
	// CancelSubscription cancels a subscription
	CancelSubscription(ctx context.Context, subscriptionID string, immediately bool) error
	
	// UpdateSubscription updates subscription (e.g., change plan)
	UpdateSubscription(ctx context.Context, sub *Subscription) error
	
	// CreatePaymentMethod adds a payment method
	CreatePaymentMethod(ctx context.Context, method *PaymentMethod) error
	
	// ChargePayment processes a one-time payment
	ChargePayment(ctx context.Context, charge *Charge) error
	
	// RefundPayment issues a refund
	RefundPayment(ctx context.Context, refund *Refund) error
	
	// ListInvoices lists customer invoices
	ListInvoices(ctx context.Context, customerID string, limit int) ([]*Invoice, error)
	
	// HandleWebhook processes provider webhooks
	HandleWebhook(ctx context.Context, payload []byte, signature string) (*WebhookEvent, error)
}

// Customer represents a customer
type Customer struct {
	ID            string                 `json:"id"`
	ProviderID    string                 `json:"provider_id"`
	Email         string                 `json:"email"`
	Name          string                 `json:"name"`
	Company       string                 `json:"company,omitempty"`
	Phone         string                 `json:"phone,omitempty"`
	Address       *Address               `json:"address,omitempty"`
	Metadata      map[string]string      `json:"metadata,omitempty"`
	PaymentMethod *PaymentMethod         `json:"payment_method,omitempty"`
	Balance       int64                  `json:"balance"` // In cents
	Currency      string                 `json:"currency"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// Address represents a billing address
type Address struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// Subscription represents a subscription
type Subscription struct {
	ID                string            `json:"id"`
	ProviderID        string            `json:"provider_id"`
	CustomerID        string            `json:"customer_id"`
	PlanID            string            `json:"plan_id"`
	Status            SubscriptionStatus `json:"status"`
	Quantity          int               `json:"quantity"`
	CurrentPeriodStart time.Time        `json:"current_period_start"`
	CurrentPeriodEnd   time.Time        `json:"current_period_end"`
	CancelAt          *time.Time        `json:"cancel_at,omitempty"`
	CanceledAt        *time.Time        `json:"canceled_at,omitempty"`
	TrialStart        *time.Time        `json:"trial_start,omitempty"`
	TrialEnd          *time.Time        `json:"trial_end,omitempty"`
	Metadata          map[string]string `json:"metadata,omitempty"`
	Items             []SubscriptionItem `json:"items,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

// SubscriptionStatus represents subscription status
type SubscriptionStatus string

const (
	StatusActive   SubscriptionStatus = "active"
	StatusTrialing SubscriptionStatus = "trialing"
	StatusPastDue  SubscriptionStatus = "past_due"
	StatusCanceled SubscriptionStatus = "canceled"
	StatusUnpaid   SubscriptionStatus = "unpaid"
	StatusPaused   SubscriptionStatus = "paused"
)

// SubscriptionItem represents a subscription line item
type SubscriptionItem struct {
	ID       string `json:"id"`
	PriceID  string `json:"price_id"`
	Quantity int    `json:"quantity"`
}

// Plan represents a subscription plan
type Plan struct {
	ID          string            `json:"id"`
	ProviderID  string            `json:"provider_id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Amount      int64             `json:"amount"` // In cents
	Currency    string            `json:"currency"`
	Interval    BillingInterval   `json:"interval"`
	Features    []string          `json:"features,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Active      bool              `json:"active"`
	TrialDays   int               `json:"trial_days"`
}

// BillingInterval represents billing frequency
type BillingInterval string

const (
	IntervalMonthly  BillingInterval = "monthly"
	IntervalYearly   BillingInterval = "yearly"
	IntervalWeekly   BillingInterval = "weekly"
	IntervalOneTime  BillingInterval = "one_time"
)

// PaymentMethod represents a payment method
type PaymentMethod struct {
	ID         string            `json:"id"`
	ProviderID string            `json:"provider_id"`
	CustomerID string            `json:"customer_id"`
	Type       PaymentMethodType `json:"type"`
	Card       *CardDetails      `json:"card,omitempty"`
	IsDefault  bool              `json:"is_default"`
	CreatedAt  time.Time         `json:"created_at"`
}

// PaymentMethodType represents payment method type
type PaymentMethodType string

const (
	PaymentCard   PaymentMethodType = "card"
	PaymentBank   PaymentMethodType = "bank"
	PaymentPayPal PaymentMethodType = "paypal"
)

// CardDetails represents credit card details
type CardDetails struct {
	Brand      string `json:"brand"`
	Last4      string `json:"last4"`
	ExpMonth   int    `json:"exp_month"`
	ExpYear    int    `json:"exp_year"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

// Charge represents a payment charge
type Charge struct {
	ID             string       `json:"id"`
	ProviderID     string       `json:"provider_id"`
	CustomerID     string       `json:"customer_id"`
	Amount         int64        `json:"amount"` // In cents
	Currency       string       `json:"currency"`
	Description    string       `json:"description"`
	Status         ChargeStatus `json:"status"`
	PaymentMethod  string       `json:"payment_method"`
	FailureMessage string       `json:"failure_message,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
}

// ChargeStatus represents charge status
type ChargeStatus string

const (
	ChargeSucceeded ChargeStatus = "succeeded"
	ChargePending   ChargeStatus = "pending"
	ChargeFailed    ChargeStatus = "failed"
)

// Refund represents a refund
type Refund struct {
	ID         string       `json:"id"`
	ProviderID string       `json:"provider_id"`
	ChargeID   string       `json:"charge_id"`
	Amount     int64        `json:"amount"` // In cents
	Currency   string       `json:"currency"`
	Reason     string       `json:"reason"`
	Status     RefundStatus `json:"status"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	CreatedAt  time.Time    `json:"created_at"`
}

// RefundStatus represents refund status
type RefundStatus string

const (
	RefundSucceeded RefundStatus = "succeeded"
	RefundPending   RefundStatus = "pending"
	RefundFailed    RefundStatus = "failed"
)

// Invoice represents an invoice
type Invoice struct {
	ID             string         `json:"id"`
	ProviderID     string         `json:"provider_id"`
	CustomerID     string         `json:"customer_id"`
	SubscriptionID string         `json:"subscription_id,omitempty"`
	Number         string         `json:"number"`
	Status         InvoiceStatus  `json:"status"`
	Amount         int64          `json:"amount"` // In cents
	Currency       string         `json:"currency"`
	DueDate        time.Time      `json:"due_date"`
	PaidAt         *time.Time     `json:"paid_at,omitempty"`
	Lines          []InvoiceLine  `json:"lines"`
	PDFUrl         string         `json:"pdf_url,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
}

// InvoiceStatus represents invoice status
type InvoiceStatus string

const (
	InvoiceDraft  InvoiceStatus = "draft"
	InvoiceOpen   InvoiceStatus = "open"
	InvoicePaid   InvoiceStatus = "paid"
	InvoiceVoid   InvoiceStatus = "void"
	InvoiceUncollectible InvoiceStatus = "uncollectible"
)

// InvoiceLine represents an invoice line item
type InvoiceLine struct {
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	UnitPrice   int64  `json:"unit_price"` // In cents
	Amount      int64  `json:"amount"`     // In cents
}

// WebhookEvent represents a webhook event
type WebhookEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
}

// Manager handles payment operations
type Manager struct {
	provider Provider
	plans    map[string]*Plan
	mu       sync.RWMutex
}

// NewManager creates a new payment manager
func NewManager(provider Provider) *Manager {
	return &Manager{
		provider: provider,
		plans:    make(map[string]*Plan),
	}
}

// CreateCustomer creates a new customer
func (m *Manager) CreateCustomer(ctx context.Context, email, name string) (*Customer, error) {
	customer := &Customer{
		Email:     email,
		Name:      name,
		Currency:  "usd",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	if err := m.provider.CreateCustomer(ctx, customer); err != nil {
		return nil, fmt.Errorf("failed to create customer: %v", err)
	}
	
	common.Info("[PAYMENT] Created customer: %s (%s)", customer.Email, customer.ID)
	return customer, nil
}

// Subscribe creates a subscription for a customer
func (m *Manager) Subscribe(ctx context.Context, customerID, planID string) (*Subscription, error) {
	// Get plan
	m.mu.RLock()
	plan, ok := m.plans[planID]
	m.mu.RUnlock()
	
	if !ok {
		return nil, fmt.Errorf("plan not found: %s", planID)
	}
	
	sub := &Subscription{
		CustomerID: customerID,
		PlanID:     planID,
		Status:     StatusActive,
		Quantity:   1,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	// Add trial if configured
	if plan.TrialDays > 0 {
		now := time.Now()
		trialEnd := now.AddDate(0, 0, plan.TrialDays)
		sub.TrialStart = &now
		sub.TrialEnd = &trialEnd
		sub.Status = StatusTrialing
	}
	
	if err := m.provider.CreateSubscription(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed to create subscription: %v", err)
	}
	
	common.Info("[PAYMENT] Created subscription: %s for customer %s", sub.ID, customerID)
	return sub, nil
}

// CancelSubscription cancels a subscription
func (m *Manager) CancelSubscription(ctx context.Context, subscriptionID string, immediately bool) error {
	if err := m.provider.CancelSubscription(ctx, subscriptionID, immediately); err != nil {
		return fmt.Errorf("failed to cancel subscription: %v", err)
	}
	
	common.Info("[PAYMENT] Canceled subscription: %s (immediately: %v)", subscriptionID, immediately)
	return nil
}

// ChangePlan changes subscription plan
func (m *Manager) ChangePlan(ctx context.Context, subscriptionID, newPlanID string) error {
	sub, err := m.provider.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %v", err)
	}
	
	sub.PlanID = newPlanID
	sub.UpdatedAt = time.Now()
	
	if err := m.provider.UpdateSubscription(ctx, sub); err != nil {
		return fmt.Errorf("failed to update subscription: %v", err)
	}
	
	common.Info("[PAYMENT] Changed subscription %s to plan %s", subscriptionID, newPlanID)
	return nil
}

// ChargeOneTime processes a one-time payment
func (m *Manager) ChargeOneTime(ctx context.Context, customerID string, amount int64, description string) (*Charge, error) {
	charge := &Charge{
		CustomerID:  customerID,
		Amount:      amount,
		Currency:    "usd",
		Description: description,
		CreatedAt:   time.Now(),
	}
	
	if err := m.provider.ChargePayment(ctx, charge); err != nil {
		return nil, fmt.Errorf("failed to charge payment: %v", err)
	}
	
	common.Info("[PAYMENT] Charged %d cents to customer %s", amount, customerID)
	return charge, nil
}

// AddPlan adds a plan to the manager
func (m *Manager) AddPlan(plan *Plan) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.plans[plan.ID] = plan
	common.Debug("[PAYMENT] Added plan: %s (%s)", plan.ID, plan.Name)
}

// GetPlan retrieves a plan
func (m *Manager) GetPlan(planID string) (*Plan, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	plan, ok := m.plans[planID]
	return plan, ok
}

// ListPlans returns all available plans
func (m *Manager) ListPlans() []*Plan {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var plans []*Plan
	for _, plan := range m.plans {
		if plan.Active {
			plans = append(plans, plan)
		}
	}
	
	return plans
}

// HandleWebhook processes payment provider webhooks
func (m *Manager) HandleWebhook(ctx context.Context, payload []byte, signature string) error {
	event, err := m.provider.HandleWebhook(ctx, payload, signature)
	if err != nil {
		return fmt.Errorf("failed to handle webhook: %v", err)
	}
	
	// Process event based on type
	switch event.Type {
	case "subscription.created":
		common.Info("[PAYMENT] Webhook: Subscription created")
	case "subscription.updated":
		common.Info("[PAYMENT] Webhook: Subscription updated")
	case "subscription.canceled":
		common.Info("[PAYMENT] Webhook: Subscription canceled")
	case "invoice.paid":
		common.Info("[PAYMENT] Webhook: Invoice paid")
	case "invoice.payment_failed":
		common.Warn("[PAYMENT] Webhook: Invoice payment failed")
	case "customer.updated":
		common.Info("[PAYMENT] Webhook: Customer updated")
	default:
		common.Debug("[PAYMENT] Webhook: Unhandled event type: %s", event.Type)
	}
	
	return nil
}

// Usage tracking

// UsageRecord represents usage data
type UsageRecord struct {
	CustomerID     string    `json:"customer_id"`
	SubscriptionID string    `json:"subscription_id"`
	Metric         string    `json:"metric"`
	Quantity       int64     `json:"quantity"`
	Timestamp      time.Time `json:"timestamp"`
}

// TrackUsage records usage for metered billing
func (m *Manager) TrackUsage(ctx context.Context, record *UsageRecord) error {
	// This would be implemented based on the payment provider's usage API
	common.Debug("[PAYMENT] Tracked usage: %s = %d for customer %s", 
		record.Metric, record.Quantity, record.CustomerID)
	return nil
}

// GetUsage retrieves usage for a period
func (m *Manager) GetUsage(ctx context.Context, customerID string, start, end time.Time) ([]*UsageRecord, error) {
	// This would query usage from the provider or database
	return []*UsageRecord{}, nil
}