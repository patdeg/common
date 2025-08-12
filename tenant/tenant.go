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

// Package tenant provides multi-tenant support with tenant isolation,
// configuration management, and cross-tenant operations.
package tenant

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/patdeg/common"
)

// Tenant represents a tenant in the system
type Tenant struct {
	ID          string                 `json:"id" datastore:"id"`
	Name        string                 `json:"name" datastore:"name"`
	Domain      string                 `json:"domain" datastore:"domain"`
	Status      TenantStatus           `json:"status" datastore:"status"`
	Plan        string                 `json:"plan" datastore:"plan"`
	Config      map[string]interface{} `json:"config" datastore:"config,noindex"`
	Limits      *TenantLimits          `json:"limits" datastore:"limits"`
	CreatedAt   time.Time              `json:"created_at" datastore:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" datastore:"updated_at"`
	TrialEndsAt *time.Time             `json:"trial_ends_at,omitempty" datastore:"trial_ends_at"`
	Metadata    map[string]string      `json:"metadata" datastore:"metadata,noindex"`
}

// TenantStatus represents the status of a tenant
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusInactive  TenantStatus = "inactive"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusTrial     TenantStatus = "trial"
	TenantStatusDeleted   TenantStatus = "deleted"
)

// TenantLimits defines resource limits for a tenant
type TenantLimits struct {
	MaxUsers        int `json:"max_users" datastore:"max_users"`
	MaxStorage      int `json:"max_storage_gb" datastore:"max_storage_gb"`
	MaxAPICallsDay  int `json:"max_api_calls_day" datastore:"max_api_calls_day"`
	MaxConcurrent   int `json:"max_concurrent" datastore:"max_concurrent"`
	CustomLimits    map[string]int `json:"custom_limits" datastore:"custom_limits,noindex"`
}

// TenantContext provides tenant-aware context
type TenantContext struct {
	context.Context
	tenant *Tenant
}

// Manager handles tenant operations
type Manager interface {
	// CreateTenant creates a new tenant
	CreateTenant(ctx context.Context, tenant *Tenant) error
	
	// GetTenant retrieves a tenant by ID
	GetTenant(ctx context.Context, tenantID string) (*Tenant, error)
	
	// GetTenantByDomain retrieves a tenant by domain
	GetTenantByDomain(ctx context.Context, domain string) (*Tenant, error)
	
	// UpdateTenant updates a tenant
	UpdateTenant(ctx context.Context, tenant *Tenant) error
	
	// DeleteTenant soft deletes a tenant
	DeleteTenant(ctx context.Context, tenantID string) error
	
	// ListTenants lists all tenants with optional filtering
	ListTenants(ctx context.Context, filter *TenantFilter) ([]*Tenant, error)
	
	// ValidateTenantAccess validates if a user has access to a tenant
	ValidateTenantAccess(ctx context.Context, tenantID string, userID string) error
	
	// GetTenantContext creates a tenant-aware context
	GetTenantContext(ctx context.Context, tenantID string) (context.Context, error)
	
	// CheckLimit checks if a tenant has exceeded a limit
	CheckLimit(ctx context.Context, tenantID string, limitName string, value int) error
}

// TenantFilter provides filtering options for listing tenants
type TenantFilter struct {
	Status   TenantStatus
	Plan     string
	Domain   string
	Limit    int
	Offset   int
}

// DefaultManager implements the Manager interface
type DefaultManager struct {
	// In production, this would use a database
	tenants map[string]*Tenant
	domains map[string]string // domain -> tenant ID
	mu      sync.RWMutex
}

// NewManager creates a new tenant manager
func NewManager() Manager {
	return &DefaultManager{
		tenants: make(map[string]*Tenant),
		domains: make(map[string]string),
	}
}

// CreateTenant creates a new tenant
func (m *DefaultManager) CreateTenant(ctx context.Context, tenant *Tenant) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Validate tenant
	if tenant.ID == "" {
		tenant.ID = generateTenantID()
	}
	
	if tenant.Name == "" {
		return fmt.Errorf("tenant name is required")
	}
	
	// Check for duplicate ID
	if _, exists := m.tenants[tenant.ID]; exists {
		return fmt.Errorf("tenant with ID %s already exists", tenant.ID)
	}
	
	// Check for duplicate domain
	if tenant.Domain != "" {
		if _, exists := m.domains[tenant.Domain]; exists {
			return fmt.Errorf("domain %s is already registered", tenant.Domain)
		}
	}
	
	// Set defaults
	if tenant.Status == "" {
		tenant.Status = TenantStatusTrial
	}
	
	if tenant.Plan == "" {
		tenant.Plan = "free"
	}
	
	if tenant.Limits == nil {
		tenant.Limits = getDefaultLimits(tenant.Plan)
	}
	
	now := time.Now()
	tenant.CreatedAt = now
	tenant.UpdatedAt = now
	
	// Set trial period if applicable
	if tenant.Status == TenantStatusTrial && tenant.TrialEndsAt == nil {
		trialEnd := now.Add(14 * 24 * time.Hour) // 14 days trial
		tenant.TrialEndsAt = &trialEnd
	}
	
	// Store tenant
	m.tenants[tenant.ID] = tenant
	if tenant.Domain != "" {
		m.domains[tenant.Domain] = tenant.ID
	}
	
	common.Info("[TENANT] Created tenant: %s (%s)", tenant.ID, tenant.Name)
	return nil
}

// GetTenant retrieves a tenant by ID
func (m *DefaultManager) GetTenant(ctx context.Context, tenantID string) (*Tenant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	tenant, exists := m.tenants[tenantID]
	if !exists {
		return nil, fmt.Errorf("tenant not found: %s", tenantID)
	}
	
	return tenant, nil
}

// GetTenantByDomain retrieves a tenant by domain
func (m *DefaultManager) GetTenantByDomain(ctx context.Context, domain string) (*Tenant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	tenantID, exists := m.domains[domain]
	if !exists {
		return nil, fmt.Errorf("no tenant found for domain: %s", domain)
	}
	
	return m.tenants[tenantID], nil
}

// UpdateTenant updates a tenant
func (m *DefaultManager) UpdateTenant(ctx context.Context, tenant *Tenant) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	existing, exists := m.tenants[tenant.ID]
	if !exists {
		return fmt.Errorf("tenant not found: %s", tenant.ID)
	}
	
	// Update domain mapping if changed
	if existing.Domain != tenant.Domain {
		if existing.Domain != "" {
			delete(m.domains, existing.Domain)
		}
		if tenant.Domain != "" {
			// Check if new domain is available
			if _, taken := m.domains[tenant.Domain]; taken {
				return fmt.Errorf("domain %s is already registered", tenant.Domain)
			}
			m.domains[tenant.Domain] = tenant.ID
		}
	}
	
	tenant.UpdatedAt = time.Now()
	m.tenants[tenant.ID] = tenant
	
	common.Info("[TENANT] Updated tenant: %s", tenant.ID)
	return nil
}

// DeleteTenant soft deletes a tenant
func (m *DefaultManager) DeleteTenant(ctx context.Context, tenantID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	tenant, exists := m.tenants[tenantID]
	if !exists {
		return fmt.Errorf("tenant not found: %s", tenantID)
	}
	
	// Soft delete - mark as deleted
	tenant.Status = TenantStatusDeleted
	tenant.UpdatedAt = time.Now()
	
	// Remove domain mapping
	if tenant.Domain != "" {
		delete(m.domains, tenant.Domain)
	}
	
	common.Info("[TENANT] Deleted tenant: %s", tenantID)
	return nil
}

// ListTenants lists all tenants with optional filtering
func (m *DefaultManager) ListTenants(ctx context.Context, filter *TenantFilter) ([]*Tenant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var results []*Tenant
	
	for _, tenant := range m.tenants {
		// Apply filters
		if filter != nil {
			if filter.Status != "" && tenant.Status != filter.Status {
				continue
			}
			if filter.Plan != "" && tenant.Plan != filter.Plan {
				continue
			}
			if filter.Domain != "" && tenant.Domain != filter.Domain {
				continue
			}
		}
		
		results = append(results, tenant)
	}
	
	// Apply pagination
	if filter != nil && filter.Limit > 0 {
		start := filter.Offset
		if start > len(results) {
			return []*Tenant{}, nil
		}
		
		end := start + filter.Limit
		if end > len(results) {
			end = len(results)
		}
		
		results = results[start:end]
	}
	
	return results, nil
}

// ValidateTenantAccess validates if a user has access to a tenant
func (m *DefaultManager) ValidateTenantAccess(ctx context.Context, tenantID string, userID string) error {
	tenant, err := m.GetTenant(ctx, tenantID)
	if err != nil {
		return err
	}
	
	// Check if tenant is active
	if tenant.Status != TenantStatusActive && tenant.Status != TenantStatusTrial {
		return fmt.Errorf("tenant is not active: %s", tenant.Status)
	}
	
	// Check trial expiration
	if tenant.Status == TenantStatusTrial && tenant.TrialEndsAt != nil {
		if time.Now().After(*tenant.TrialEndsAt) {
			return fmt.Errorf("trial period has expired")
		}
	}
	
	// In production, check user-tenant association
	// For now, we'll assume access is granted
	
	return nil
}

// GetTenantContext creates a tenant-aware context
func (m *DefaultManager) GetTenantContext(ctx context.Context, tenantID string) (context.Context, error) {
	tenant, err := m.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	
	return WithTenant(ctx, tenant), nil
}

// CheckLimit checks if a tenant has exceeded a limit
func (m *DefaultManager) CheckLimit(ctx context.Context, tenantID string, limitName string, value int) error {
	tenant, err := m.GetTenant(ctx, tenantID)
	if err != nil {
		return err
	}
	
	if tenant.Limits == nil {
		return nil // No limits configured
	}
	
	var limit int
	switch limitName {
	case "users":
		limit = tenant.Limits.MaxUsers
	case "storage":
		limit = tenant.Limits.MaxStorage
	case "api_calls":
		limit = tenant.Limits.MaxAPICallsDay
	case "concurrent":
		limit = tenant.Limits.MaxConcurrent
	default:
		if tenant.Limits.CustomLimits != nil {
			limit = tenant.Limits.CustomLimits[limitName]
		}
	}
	
	if limit > 0 && value > limit {
		return fmt.Errorf("tenant limit exceeded: %s (current: %d, limit: %d)", limitName, value, limit)
	}
	
	return nil
}

// Context functions

// WithTenant adds a tenant to the context
func WithTenant(ctx context.Context, tenant *Tenant) context.Context {
	return &TenantContext{
		Context: ctx,
		tenant:  tenant,
	}
}

// FromContext retrieves a tenant from the context
func FromContext(ctx context.Context) (*Tenant, bool) {
	if tc, ok := ctx.(*TenantContext); ok {
		return tc.tenant, true
	}
	return nil, false
}

// Helper functions

func generateTenantID() string {
	return fmt.Sprintf("tenant_%d", time.Now().UnixNano())
}

func getDefaultLimits(plan string) *TenantLimits {
	switch plan {
	case "enterprise":
		return &TenantLimits{
			MaxUsers:       -1, // Unlimited
			MaxStorage:     1000,
			MaxAPICallsDay: 1000000,
			MaxConcurrent:  100,
		}
	case "pro":
		return &TenantLimits{
			MaxUsers:       100,
			MaxStorage:     100,
			MaxAPICallsDay: 100000,
			MaxConcurrent:  50,
		}
	case "starter":
		return &TenantLimits{
			MaxUsers:       10,
			MaxStorage:     10,
			MaxAPICallsDay: 10000,
			MaxConcurrent:  10,
		}
	default: // free
		return &TenantLimits{
			MaxUsers:       5,
			MaxStorage:     1,
			MaxAPICallsDay: 1000,
			MaxConcurrent:  5,
		}
	}
}

// TenantMiddleware provides HTTP middleware for tenant extraction
func TenantMiddleware(manager Manager) func(next func(context.Context) error) func(context.Context) error {
	return func(next func(context.Context) error) func(context.Context) error {
		return func(ctx context.Context) error {
			// Extract tenant from context (e.g., from domain, header, or path)
			// This is a simplified example
			
			// For now, just pass through
			return next(ctx)
		}
	}
}