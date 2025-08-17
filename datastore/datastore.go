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

// Package datastore provides a flexible data storage abstraction that supports
// both local development (in-memory) and production (Google Cloud Datastore).
package datastore

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/patdeg/common"
)

// Repository defines the generic repository interface for entity operations
type Repository interface {
	// Get retrieves an entity by key
	Get(ctx context.Context, kind string, key string, dest interface{}) error

	// Put saves an entity with the given key
	Put(ctx context.Context, kind string, key string, src interface{}) error

	// Delete removes an entity by key
	Delete(ctx context.Context, kind string, key string) error

	// Query executes a query and returns results
	Query(ctx context.Context, query Query) ([]interface{}, error)

	// Transaction executes operations in a transaction
	Transaction(ctx context.Context, fn func(tx Transaction) error) error
}

// Transaction represents a transactional context
type Transaction interface {
	Get(kind string, key string, dest interface{}) error
	Put(kind string, key string, src interface{}) error
	Delete(kind string, key string) error
}

// Query represents a datastore query
type Query struct {
	Kind    string
	Filters []Filter
	Orders  []Order
	Limit   int
	Offset  int
}

// Filter represents a query filter
type Filter struct {
	Field    string
	Operator string
	Value    interface{}
}

// Order represents a query ordering
type Order struct {
	Field      string
	Descending bool
}

// CloudRepository implements Repository using Google Cloud Datastore
type CloudRepository struct {
	client    *datastore.Client
	projectID string
	cache     sync.Map // Local cache for frequently accessed items
}

// LocalRepository implements Repository using in-memory storage for development
type LocalRepository struct {
	data map[string]map[string]interface{} // kind -> key -> entity
	mu   sync.RWMutex
}

// NewRepository creates a new repository based on the environment
func NewRepository(ctx context.Context) (Repository, error) {
	if isLocalDevelopment() {
		common.Info("[DATASTORE] Initializing LocalRepository for development")
		return NewLocalRepository(), nil
	}

	common.Info("[DATASTORE] Initializing CloudRepository for production")
	return NewCloudRepository(ctx)
}

// NewCloudRepository creates a new cloud-based repository
func NewCloudRepository(ctx context.Context) (*CloudRepository, error) {
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}
	if projectID == "" {
		return nil, fmt.Errorf("PROJECT_ID not configured")
	}

	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create datastore client: %v", err)
	}

	return &CloudRepository{
		client:    client,
		projectID: projectID,
	}, nil
}

// NewLocalRepository creates a new local in-memory repository
func NewLocalRepository() *LocalRepository {
	return &LocalRepository{
		data: make(map[string]map[string]interface{}),
	}
}

// Get retrieves an entity from cloud datastore
func (r *CloudRepository) Get(ctx context.Context, kind string, key string, dest interface{}) error {
	// Check cache first
	cacheKey := fmt.Sprintf("%s:%s", kind, key)
	if cached, ok := r.cache.Load(cacheKey); ok {
		// Type assertion and copy
		return copyValue(cached, dest)
	}

	k := datastore.NameKey(kind, key, nil)
	err := r.client.Get(ctx, k, dest)
	if err != nil {
		return err
	}

	// Cache the result
	r.cache.Store(cacheKey, dest)
	return nil
}

// Put saves an entity to cloud datastore
func (r *CloudRepository) Put(ctx context.Context, kind string, key string, src interface{}) error {
	k := datastore.NameKey(kind, key, nil)
	_, err := r.client.Put(ctx, k, src)
	if err != nil {
		return err
	}

	// Update cache
	cacheKey := fmt.Sprintf("%s:%s", kind, key)
	r.cache.Store(cacheKey, src)
	return nil
}

// Delete removes an entity from cloud datastore
func (r *CloudRepository) Delete(ctx context.Context, kind string, key string) error {
	k := datastore.NameKey(kind, key, nil)
	err := r.client.Delete(ctx, k)
	if err != nil {
		return err
	}

	// Remove from cache
	cacheKey := fmt.Sprintf("%s:%s", kind, key)
	r.cache.Delete(cacheKey)
	return nil
}

// Query executes a query on cloud datastore
func (r *CloudRepository) Query(ctx context.Context, query Query) ([]interface{}, error) {
	q := datastore.NewQuery(query.Kind)

	// Apply filters
	for _, filter := range query.Filters {
		q = q.Filter(fmt.Sprintf("%s %s", filter.Field, filter.Operator), filter.Value)
	}

	// Apply ordering
	for _, order := range query.Orders {
		field := order.Field
		if order.Descending {
			field = "-" + field
		}
		q = q.Order(field)
	}

	// Apply limit and offset
	if query.Limit > 0 {
		q = q.Limit(query.Limit)
	}
	if query.Offset > 0 {
		q = q.Offset(query.Offset)
	}

	var results []interface{}
	_, err := r.client.GetAll(ctx, q, &results)
	return results, err
}

// Transaction executes operations in a cloud datastore transaction
func (r *CloudRepository) Transaction(ctx context.Context, fn func(tx Transaction) error) error {
	_, err := r.client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		return fn(&cloudTransaction{tx: tx})
	})
	return err
}

// cloudTransaction wraps a datastore transaction
type cloudTransaction struct {
	tx *datastore.Transaction
}

func (t *cloudTransaction) Get(kind string, key string, dest interface{}) error {
	k := datastore.NameKey(kind, key, nil)
	return t.tx.Get(k, dest)
}

func (t *cloudTransaction) Put(kind string, key string, src interface{}) error {
	k := datastore.NameKey(kind, key, nil)
	_, err := t.tx.Put(k, src)
	return err
}

func (t *cloudTransaction) Delete(kind string, key string) error {
	k := datastore.NameKey(kind, key, nil)
	return t.tx.Delete(k)
}

// Get retrieves an entity from local storage
func (r *LocalRepository) Get(ctx context.Context, kind string, key string, dest interface{}) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	kindData, ok := r.data[kind]
	if !ok {
		return fmt.Errorf("entity not found")
	}

	entity, ok := kindData[key]
	if !ok {
		return fmt.Errorf("entity not found")
	}

	return copyValue(entity, dest)
}

// Put saves an entity to local storage
func (r *LocalRepository) Put(ctx context.Context, kind string, key string, src interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.data[kind] == nil {
		r.data[kind] = make(map[string]interface{})
	}

	// Store a copy to prevent external modifications
	r.data[kind][key] = deepCopy(src)
	return nil
}

// Delete removes an entity from local storage
func (r *LocalRepository) Delete(ctx context.Context, kind string, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if kindData, ok := r.data[kind]; ok {
		delete(kindData, key)
	}
	return nil
}

// Query executes a query on local storage
func (r *LocalRepository) Query(ctx context.Context, query Query) ([]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	kindData, ok := r.data[query.Kind]
	if !ok {
		return []interface{}{}, nil
	}

	// Collect all entities
	var results []interface{}
	for _, entity := range kindData {
		// TODO: Apply filters, ordering, limit, offset
		// This is a simplified implementation
		results = append(results, entity)
		if query.Limit > 0 && len(results) >= query.Limit {
			break
		}
	}

	return results, nil
}

// Transaction executes operations in a local transaction (simplified)
func (r *LocalRepository) Transaction(ctx context.Context, fn func(tx Transaction) error) error {
	// Simplified transaction for local storage
	// In production, this would need proper isolation
	return fn(&localTransaction{repo: r})
}

// localTransaction wraps local repository operations
type localTransaction struct {
	repo *LocalRepository
}

func (t *localTransaction) Get(kind string, key string, dest interface{}) error {
	return t.repo.Get(context.Background(), kind, key, dest)
}

func (t *localTransaction) Put(kind string, key string, src interface{}) error {
	return t.repo.Put(context.Background(), kind, key, src)
}

func (t *localTransaction) Delete(kind string, key string) error {
	return t.repo.Delete(context.Background(), kind, key)
}

// Helper functions

func isLocalDevelopment() bool {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = os.Getenv("GAE_ENV")
	}
	return env == "" || env == "development" || env == "local"
}

func copyValue(src, dest interface{}) error {
	// Use JSON encoding/decoding for deep copy
	// This ensures cached data is properly isolated from the original
	data, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("failed to marshal source: %w", err)
	}
	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("failed to unmarshal to destination: %w", err)
	}
	return nil
}

func deepCopy(src interface{}) interface{} {
	// Use JSON for deep copy to ensure isolation
	data, err := json.Marshal(src)
	if err != nil {
		// If marshaling fails, return the original (best effort)
		return src
	}
	
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		// If unmarshaling fails, return the original (best effort)
		return src
	}
	return result
}

// BaseEntity provides common fields for all entities
type BaseEntity struct {
	CreatedAt time.Time `datastore:"created_at"`
	UpdatedAt time.Time `datastore:"updated_at"`
	Version   int       `datastore:"version"`
}

// BeforeSave is called before saving an entity
func (e *BaseEntity) BeforeSave() {
	now := time.Now()
	if e.CreatedAt.IsZero() {
		e.CreatedAt = now
	}
	e.UpdatedAt = now
	e.Version++
}
