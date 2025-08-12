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

// Package rbac provides role-based access control with permissions,
// roles, and policy management.
package rbac

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/patdeg/common"
)

// Permission represents a single permission
type Permission struct {
	ID          string `json:"id" datastore:"id"`
	Name        string `json:"name" datastore:"name"`
	Resource    string `json:"resource" datastore:"resource"`
	Action      string `json:"action" datastore:"action"`
	Description string `json:"description" datastore:"description,noindex"`
}

// Role represents a role with a set of permissions
type Role struct {
	ID          string       `json:"id" datastore:"id"`
	Name        string       `json:"name" datastore:"name"`
	Description string       `json:"description" datastore:"description,noindex"`
	Permissions []Permission `json:"permissions" datastore:"permissions"`
	IsSystem    bool         `json:"is_system" datastore:"is_system"`
	TenantID    string       `json:"tenant_id" datastore:"tenant_id"`
	CreatedAt   time.Time    `json:"created_at" datastore:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" datastore:"updated_at"`
}

// UserRole represents the assignment of a role to a user
type UserRole struct {
	UserID    string     `json:"user_id" datastore:"user_id"`
	RoleID    string     `json:"role_id" datastore:"role_id"`
	TenantID  string     `json:"tenant_id" datastore:"tenant_id"`
	GrantedBy string     `json:"granted_by" datastore:"granted_by"`
	GrantedAt time.Time  `json:"granted_at" datastore:"granted_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty" datastore:"expires_at"`
}

// Policy represents an access control policy
type Policy struct {
	ID          string                 `json:"id" datastore:"id"`
	Name        string                 `json:"name" datastore:"name"`
	Description string                 `json:"description" datastore:"description,noindex"`
	Rules       []PolicyRule           `json:"rules" datastore:"rules"`
	TenantID    string                 `json:"tenant_id" datastore:"tenant_id"`
	Priority    int                    `json:"priority" datastore:"priority"`
	Enabled     bool                   `json:"enabled" datastore:"enabled"`
	Conditions  map[string]interface{} `json:"conditions" datastore:"conditions,noindex"`
}

// PolicyRule represents a single rule in a policy
type PolicyRule struct {
	Resource   string   `json:"resource"`
	Actions    []string `json:"actions"`
	Effect     Effect   `json:"effect"`
	Principals []string `json:"principals"` // User IDs or role IDs
}

// Effect represents the effect of a policy rule
type Effect string

const (
	EffectAllow Effect = "allow"
	EffectDeny  Effect = "deny"
)

// Manager handles RBAC operations
type Manager interface {
	// Role management
	CreateRole(ctx context.Context, role *Role) error
	GetRole(ctx context.Context, roleID string) (*Role, error)
	UpdateRole(ctx context.Context, role *Role) error
	DeleteRole(ctx context.Context, roleID string) error
	ListRoles(ctx context.Context, tenantID string) ([]*Role, error)

	// User-Role assignment
	AssignRole(ctx context.Context, userID, roleID, tenantID string) error
	RevokeRole(ctx context.Context, userID, roleID, tenantID string) error
	GetUserRoles(ctx context.Context, userID, tenantID string) ([]*Role, error)
	HasRole(ctx context.Context, userID, roleID, tenantID string) bool

	// Permission checking
	HasPermission(ctx context.Context, userID, resource, action, tenantID string) bool
	GetUserPermissions(ctx context.Context, userID, tenantID string) ([]Permission, error)

	// Policy management
	CreatePolicy(ctx context.Context, policy *Policy) error
	GetPolicy(ctx context.Context, policyID string) (*Policy, error)
	UpdatePolicy(ctx context.Context, policy *Policy) error
	DeletePolicy(ctx context.Context, policyID string) error
	EvaluatePolicy(ctx context.Context, userID, resource, action, tenantID string) Effect
}

// DefaultManager implements the Manager interface
type DefaultManager struct {
	roles       map[string]*Role
	userRoles   map[string][]*UserRole // userID -> roles
	policies    map[string]*Policy
	permissions map[string]*Permission
	mu          sync.RWMutex
}

// NewManager creates a new RBAC manager
func NewManager() Manager {
	m := &DefaultManager{
		roles:       make(map[string]*Role),
		userRoles:   make(map[string][]*UserRole),
		policies:    make(map[string]*Policy),
		permissions: make(map[string]*Permission),
	}

	// Initialize with default roles
	m.initializeDefaultRoles()

	return m
}

// initializeDefaultRoles creates system default roles
func (m *DefaultManager) initializeDefaultRoles() {
	// Admin role
	adminRole := &Role{
		ID:          "admin",
		Name:        "Administrator",
		Description: "Full system access",
		IsSystem:    true,
		Permissions: []Permission{
			{ID: "all", Name: "All Permissions", Resource: "*", Action: "*"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.roles[adminRole.ID] = adminRole

	// User role
	userRole := &Role{
		ID:          "user",
		Name:        "User",
		Description: "Standard user access",
		IsSystem:    true,
		Permissions: []Permission{
			{ID: "read_own", Name: "Read Own Data", Resource: "user:self", Action: "read"},
			{ID: "write_own", Name: "Write Own Data", Resource: "user:self", Action: "write"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.roles[userRole.ID] = userRole

	// Viewer role
	viewerRole := &Role{
		ID:          "viewer",
		Name:        "Viewer",
		Description: "Read-only access",
		IsSystem:    true,
		Permissions: []Permission{
			{ID: "read_all", Name: "Read All Data", Resource: "*", Action: "read"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.roles[viewerRole.ID] = viewerRole
}

// CreateRole creates a new role
func (m *DefaultManager) CreateRole(ctx context.Context, role *Role) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if role.ID == "" {
		role.ID = fmt.Sprintf("role_%d", time.Now().UnixNano())
	}

	if _, exists := m.roles[role.ID]; exists {
		return fmt.Errorf("role already exists: %s", role.ID)
	}

	now := time.Now()
	role.CreatedAt = now
	role.UpdatedAt = now

	m.roles[role.ID] = role

	common.Info("[RBAC] Created role: %s (%s)", role.ID, role.Name)
	return nil
}

// GetRole retrieves a role by ID
func (m *DefaultManager) GetRole(ctx context.Context, roleID string) (*Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	role, exists := m.roles[roleID]
	if !exists {
		return nil, fmt.Errorf("role not found: %s", roleID)
	}

	return role, nil
}

// UpdateRole updates an existing role
func (m *DefaultManager) UpdateRole(ctx context.Context, role *Role) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, exists := m.roles[role.ID]
	if !exists {
		return fmt.Errorf("role not found: %s", role.ID)
	}

	if existing.IsSystem {
		return fmt.Errorf("cannot modify system role: %s", role.ID)
	}

	role.UpdatedAt = time.Now()
	m.roles[role.ID] = role

	common.Info("[RBAC] Updated role: %s", role.ID)
	return nil
}

// DeleteRole deletes a role
func (m *DefaultManager) DeleteRole(ctx context.Context, roleID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	role, exists := m.roles[roleID]
	if !exists {
		return fmt.Errorf("role not found: %s", roleID)
	}

	if role.IsSystem {
		return fmt.Errorf("cannot delete system role: %s", roleID)
	}

	// Remove all user assignments for this role
	for userID, userRoles := range m.userRoles {
		var filtered []*UserRole
		for _, ur := range userRoles {
			if ur.RoleID != roleID {
				filtered = append(filtered, ur)
			}
		}
		m.userRoles[userID] = filtered
	}

	delete(m.roles, roleID)

	common.Info("[RBAC] Deleted role: %s", roleID)
	return nil
}

// ListRoles lists all roles for a tenant
func (m *DefaultManager) ListRoles(ctx context.Context, tenantID string) ([]*Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var roles []*Role
	for _, role := range m.roles {
		if role.TenantID == tenantID || role.TenantID == "" || role.IsSystem {
			roles = append(roles, role)
		}
	}

	return roles, nil
}

// AssignRole assigns a role to a user
func (m *DefaultManager) AssignRole(ctx context.Context, userID, roleID, tenantID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if role exists
	if _, exists := m.roles[roleID]; !exists {
		return fmt.Errorf("role not found: %s", roleID)
	}

	// Check if already assigned
	for _, ur := range m.userRoles[userID] {
		if ur.RoleID == roleID && ur.TenantID == tenantID {
			return fmt.Errorf("role already assigned")
		}
	}

	userRole := &UserRole{
		UserID:    userID,
		RoleID:    roleID,
		TenantID:  tenantID,
		GrantedAt: time.Now(),
	}

	m.userRoles[userID] = append(m.userRoles[userID], userRole)

	common.Info("[RBAC] Assigned role %s to user %s", roleID, userID)
	return nil
}

// RevokeRole revokes a role from a user
func (m *DefaultManager) RevokeRole(ctx context.Context, userID, roleID, tenantID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var filtered []*UserRole
	found := false

	for _, ur := range m.userRoles[userID] {
		if ur.RoleID == roleID && ur.TenantID == tenantID {
			found = true
		} else {
			filtered = append(filtered, ur)
		}
	}

	if !found {
		return fmt.Errorf("role assignment not found")
	}

	m.userRoles[userID] = filtered

	common.Info("[RBAC] Revoked role %s from user %s", roleID, userID)
	return nil
}

// GetUserRoles gets all roles assigned to a user
func (m *DefaultManager) GetUserRoles(ctx context.Context, userID, tenantID string) ([]*Role, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var roles []*Role

	for _, ur := range m.userRoles[userID] {
		if ur.TenantID == tenantID {
			// Check if not expired
			if ur.ExpiresAt != nil && time.Now().After(*ur.ExpiresAt) {
				continue
			}

			if role, exists := m.roles[ur.RoleID]; exists {
				roles = append(roles, role)
			}
		}
	}

	return roles, nil
}

// HasRole checks if a user has a specific role
func (m *DefaultManager) HasRole(ctx context.Context, userID, roleID, tenantID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, ur := range m.userRoles[userID] {
		if ur.RoleID == roleID && ur.TenantID == tenantID {
			// Check if not expired
			if ur.ExpiresAt != nil && time.Now().After(*ur.ExpiresAt) {
				return false
			}
			return true
		}
	}

	return false
}

// HasPermission checks if a user has a specific permission
func (m *DefaultManager) HasPermission(ctx context.Context, userID, resource, action, tenantID string) bool {
	// First check policies
	effect := m.EvaluatePolicy(ctx, userID, resource, action, tenantID)
	if effect == EffectDeny {
		return false
	}
	if effect == EffectAllow {
		return true
	}

	// Then check role-based permissions
	roles, _ := m.GetUserRoles(ctx, userID, tenantID)

	for _, role := range roles {
		for _, perm := range role.Permissions {
			if matchesResource(perm.Resource, resource) && matchesAction(perm.Action, action) {
				return true
			}
		}
	}

	return false
}

// GetUserPermissions gets all permissions for a user
func (m *DefaultManager) GetUserPermissions(ctx context.Context, userID, tenantID string) ([]Permission, error) {
	roles, err := m.GetUserRoles(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}

	permMap := make(map[string]Permission)

	for _, role := range roles {
		for _, perm := range role.Permissions {
			permMap[perm.ID] = perm
		}
	}

	var permissions []Permission
	for _, perm := range permMap {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// CreatePolicy creates a new policy
func (m *DefaultManager) CreatePolicy(ctx context.Context, policy *Policy) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if policy.ID == "" {
		policy.ID = fmt.Sprintf("policy_%d", time.Now().UnixNano())
	}

	if _, exists := m.policies[policy.ID]; exists {
		return fmt.Errorf("policy already exists: %s", policy.ID)
	}

	m.policies[policy.ID] = policy

	common.Info("[RBAC] Created policy: %s (%s)", policy.ID, policy.Name)
	return nil
}

// GetPolicy retrieves a policy by ID
func (m *DefaultManager) GetPolicy(ctx context.Context, policyID string) (*Policy, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	policy, exists := m.policies[policyID]
	if !exists {
		return nil, fmt.Errorf("policy not found: %s", policyID)
	}

	return policy, nil
}

// UpdatePolicy updates an existing policy
func (m *DefaultManager) UpdatePolicy(ctx context.Context, policy *Policy) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.policies[policy.ID]; !exists {
		return fmt.Errorf("policy not found: %s", policy.ID)
	}

	m.policies[policy.ID] = policy

	common.Info("[RBAC] Updated policy: %s", policy.ID)
	return nil
}

// DeletePolicy deletes a policy
func (m *DefaultManager) DeletePolicy(ctx context.Context, policyID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.policies[policyID]; !exists {
		return fmt.Errorf("policy not found: %s", policyID)
	}

	delete(m.policies, policyID)

	common.Info("[RBAC] Deleted policy: %s", policyID)
	return nil
}

// EvaluatePolicy evaluates policies for a user action
func (m *DefaultManager) EvaluatePolicy(ctx context.Context, userID, resource, action, tenantID string) Effect {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Get user's roles
	var userRoleIDs []string
	for _, ur := range m.userRoles[userID] {
		if ur.TenantID == tenantID {
			userRoleIDs = append(userRoleIDs, ur.RoleID)
		}
	}

	// Evaluate policies in priority order
	var effect Effect

	for _, policy := range m.policies {
		if !policy.Enabled || policy.TenantID != tenantID {
			continue
		}

		for _, rule := range policy.Rules {
			// Check if rule applies to this resource and action
			if !matchesResource(rule.Resource, resource) {
				continue
			}

			actionMatches := false
			for _, a := range rule.Actions {
				if matchesAction(a, action) {
					actionMatches = true
					break
				}
			}
			if !actionMatches {
				continue
			}

			// Check if rule applies to this user
			principalMatches := false
			for _, principal := range rule.Principals {
				if principal == userID || principal == "*" {
					principalMatches = true
					break
				}
				// Check if principal is a role
				for _, roleID := range userRoleIDs {
					if principal == "role:"+roleID {
						principalMatches = true
						break
					}
				}
			}

			if principalMatches {
				effect = rule.Effect
				// Deny takes precedence
				if effect == EffectDeny {
					return EffectDeny
				}
			}
		}
	}

	return effect
}

// Helper functions

func matchesResource(pattern, resource string) bool {
	if pattern == "*" {
		return true
	}
	if pattern == resource {
		return true
	}
	// Support wildcard matching
	if strings.HasSuffix(pattern, "*") {
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(resource, prefix)
	}
	return false
}

func matchesAction(pattern, action string) bool {
	if pattern == "*" {
		return true
	}
	return pattern == action
}

// StandardRoles provides standard role definitions
var StandardRoles = struct {
	Admin  string
	User   string
	Viewer string
	Editor string
	Owner  string
}{
	Admin:  "admin",
	User:   "user",
	Viewer: "viewer",
	Editor: "editor",
	Owner:  "owner",
}

// StandardPermissions provides standard permission definitions
var StandardPermissions = struct {
	Read   string
	Write  string
	Delete string
	Admin  string
}{
	Read:   "read",
	Write:  "write",
	Delete: "delete",
	Admin:  "admin",
}
