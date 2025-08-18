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

// Package bigquery provides utilities for BigQuery operations including
// streaming inserts, table management, and query execution.
package bigquery

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/patdeg/common"
	"google.golang.org/api/googleapi"
)

// Client provides efficient BigQuery operations with lazy table creation
type Client struct {
	projectID string
	datasetID string
	client    *bigquery.Client

	// Track which tables we've already verified/created in this process
	verifiedTables map[string]bool
	mu             sync.RWMutex

	// Batch insert support
	batchSize     int
	batchInterval time.Duration
	batches       map[string][]interface{}
	batchMu       sync.Mutex
	stopBatch     chan struct{}
}

// Config contains configuration for BigQuery client
type Config struct {
	ProjectID     string
	DatasetID     string
	BatchSize     int           // Number of rows to batch before inserting
	BatchInterval time.Duration // Max time to wait before flushing batch
}

// NewClient creates a new BigQuery client
func NewClient(ctx context.Context, config Config) (*Client, error) {
	if config.ProjectID == "" {
		config.ProjectID = os.Getenv("PROJECT_ID")
		if config.ProjectID == "" {
			config.ProjectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
		}
	}

	if config.DatasetID == "" {
		config.DatasetID = os.Getenv("BQ_DATASET")
	}

	if config.ProjectID == "" || config.DatasetID == "" {
		return nil, fmt.Errorf("PROJECT_ID and DATASET_ID are required")
	}

	if config.BatchSize == 0 {
		config.BatchSize = 100
	}

	if config.BatchInterval == 0 {
		config.BatchInterval = 5 * time.Second
	}

	client, err := bigquery.NewClient(ctx, config.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create BigQuery client: %v", err)
	}

	c := &Client{
		projectID:      config.ProjectID,
		datasetID:      config.DatasetID,
		client:         client,
		verifiedTables: make(map[string]bool),
		batchSize:      config.BatchSize,
		batchInterval:  config.BatchInterval,
		batches:        make(map[string][]interface{}),
		stopBatch:      make(chan struct{}),
	}

	// Start batch processor
	go c.processBatches()

	return c, nil
}

// Close closes the BigQuery client and flushes pending batches
func (c *Client) Close(ctx context.Context) error {
	// Stop batch processor
	close(c.stopBatch)

	// Flush all pending batches
	c.flushAllBatches(ctx)

	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// InsertRow inserts a single row into a BigQuery table
func (c *Client) InsertRow(ctx context.Context, tableID string, row interface{}, schema bigquery.Schema) error {
	// Check if table has been verified
	c.mu.RLock()
	verified := c.verifiedTables[tableID]
	c.mu.RUnlock()

	if !verified {
		// Try optimistic insert first
		err := c.tryInsert(ctx, tableID, []interface{}{row})
		if err == nil {
			c.markTableVerified(tableID)
			return nil
		}

		// Check if error is because table doesn't exist
		if !isTableNotFoundError(err) {
			return fmt.Errorf("insert failed: %v", err)
		}

		// Create table
		common.Info("[BQ] Creating table %s.%s", c.datasetID, tableID)
		if err := c.ensureTableExists(ctx, tableID, schema); err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}

		// Wait for table to be ready
		time.Sleep(1 * time.Second)

		// Try insert again
		if err := c.tryInsert(ctx, tableID, []interface{}{row}); err != nil {
			return fmt.Errorf("insert failed after table creation: %v", err)
		}

		c.markTableVerified(tableID)
		return nil
	}

	// Table already verified
	return c.tryInsert(ctx, tableID, []interface{}{row})
}

// InsertRowAsync adds a row to the batch for async insertion
func (c *Client) InsertRowAsync(tableID string, row interface{}) {
	c.batchMu.Lock()
	defer c.batchMu.Unlock()

	c.batches[tableID] = append(c.batches[tableID], row)

	// Check if batch is full
	if len(c.batches[tableID]) >= c.batchSize {
		// Flush this table's batch
		rows := c.batches[tableID]
		c.batches[tableID] = nil

		// Insert in background
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := c.tryInsert(ctx, tableID, rows); err != nil {
				common.Error("[BQ] Batch insert failed for table %s: %v", tableID, err)
			}
		}()
	}
}

// Query executes a BigQuery SQL query
func (c *Client) Query(ctx context.Context, sql string, params ...bigquery.QueryParameter) (*bigquery.RowIterator, error) {
	q := c.client.Query(sql)
	if len(params) > 0 {
		q.Parameters = params
	}

	return q.Read(ctx)
}

// GetDataset returns the dataset reference
func (c *Client) GetDataset() *bigquery.Dataset {
	return c.client.Dataset(c.datasetID)
}

// GetTable returns a table reference
func (c *Client) GetTable(tableID string) *bigquery.Table {
	return c.GetDataset().Table(tableID)
}

// Private methods

func (c *Client) tryInsert(ctx context.Context, tableID string, rows []interface{}) error {
	table := c.GetTable(tableID)
	inserter := table.Inserter()
	return inserter.Put(ctx, rows)
}

func (c *Client) ensureTableExists(ctx context.Context, tableID string, schema bigquery.Schema) error {
	dataset := c.GetDataset()

	// Check if dataset exists
	if _, err := dataset.Metadata(ctx); err != nil {
		// Create dataset
		if err := dataset.Create(ctx, &bigquery.DatasetMetadata{
			Location: "US",
		}); err != nil {
			return fmt.Errorf("failed to create dataset: %v", err)
		}
	}

	// Create table
	table := dataset.Table(tableID)
	if err := table.Create(ctx, &bigquery.TableMetadata{
		Schema: schema,
	}); err != nil {
		// Check if table already exists (race condition)
		if !isAlreadyExistsError(err) {
			return fmt.Errorf("failed to create table: %v", err)
		}
	}

	return nil
}

func (c *Client) markTableVerified(tableID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.verifiedTables[tableID] = true
}

func (c *Client) processBatches() {
	ticker := time.NewTicker(c.batchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.flushAllBatches(context.Background())
		case <-c.stopBatch:
			return
		}
	}
}

func (c *Client) flushAllBatches(ctx context.Context) {
	c.batchMu.Lock()
	defer c.batchMu.Unlock()

	for tableID, rows := range c.batches {
		if len(rows) > 0 {
			// Clear batch
			c.batches[tableID] = nil

			// Insert in background
			go func(table string, data []interface{}) {
				insertCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
				defer cancel()

				if err := c.tryInsert(insertCtx, table, data); err != nil {
					common.Error("[BQ] Batch flush failed for table %s: %v", table, err)
				} else {
					common.Debug("[BQ] Flushed %d rows to table %s", len(data), table)
				}
			}(tableID, rows)
		}
	}
}

// Helper functions

func isTableNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific BigQuery error
	if e, ok := err.(*googleapi.Error); ok {
		return e.Code == 404
	}

	// Also check error message
	errStr := err.Error()
	return strings.Contains(errStr, "not found") || strings.Contains(errStr, "does not exist")
}

func isAlreadyExistsError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific BigQuery error
	if e, ok := err.(*googleapi.Error); ok {
		return e.Code == 409
	}

	// Also check error message
	errStr := err.Error()
	return strings.Contains(errStr, "already exists")
}

// SchemaFromStruct generates a BigQuery schema from a struct
// The struct should have `bigquery` tags
func SchemaFromStruct(v interface{}) (bigquery.Schema, error) {
	// This is a simplified version
	// In production, use reflection to generate schema from struct tags
	return bigquery.InferSchema(v)
}

// StandardSchemas provides common BigQuery schemas
var StandardSchemas = struct {
	Telemetry bigquery.Schema
	Audit     bigquery.Schema
	Analytics bigquery.Schema
}{
	Telemetry: bigquery.Schema{
		{Name: "timestamp", Type: bigquery.TimestampFieldType, Required: true},
		{Name: "user_id", Type: bigquery.StringFieldType},
		{Name: "session_id", Type: bigquery.StringFieldType},
		{Name: "event_type", Type: bigquery.StringFieldType, Required: true},
		{Name: "event_data", Type: bigquery.JSONFieldType},
		{Name: "duration_ms", Type: bigquery.IntegerFieldType},
		{Name: "error", Type: bigquery.StringFieldType},
	},
	Audit: bigquery.Schema{
		{Name: "timestamp", Type: bigquery.TimestampFieldType, Required: true},
		{Name: "user_id", Type: bigquery.StringFieldType, Required: true},
		{Name: "action", Type: bigquery.StringFieldType, Required: true},
		{Name: "resource", Type: bigquery.StringFieldType},
		{Name: "result", Type: bigquery.StringFieldType},
		{Name: "ip_address", Type: bigquery.StringFieldType},
		{Name: "user_agent", Type: bigquery.StringFieldType},
	},
	Analytics: bigquery.Schema{
		{Name: "timestamp", Type: bigquery.TimestampFieldType, Required: true},
		{Name: "metric_name", Type: bigquery.StringFieldType, Required: true},
		{Name: "metric_value", Type: bigquery.FloatFieldType, Required: true},
		{Name: "dimensions", Type: bigquery.JSONFieldType},
		{Name: "user_id", Type: bigquery.StringFieldType},
		{Name: "session_id", Type: bigquery.StringFieldType},
	},
}
