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

// Package tasks provides utilities for Google Cloud Tasks operations
// including task creation, queue management, and retry policies.
package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/patdeg/common"
)

// TaskQueue provides an interface for task queue operations
type TaskQueue interface {
	// CreateTask creates a new task in the queue
	CreateTask(ctx context.Context, task *Task) error
	
	// CreateHTTPTask creates an HTTP task
	CreateHTTPTask(ctx context.Context, queueName string, url string, payload interface{}) error
	
	// DeleteTask deletes a task from the queue
	DeleteTask(ctx context.Context, queueName string, taskName string) error
	
	// PurgeQueue purges all tasks from a queue
	PurgeQueue(ctx context.Context, queueName string) error
	
	// GetQueueStats returns statistics for a queue
	GetQueueStats(ctx context.Context, queueName string) (*QueueStats, error)
}

// Task represents a task to be executed
type Task struct {
	Name        string            // Task name (optional, will be generated if empty)
	Queue       string            // Queue name
	URL         string            // Target URL for HTTP tasks
	Method      string            // HTTP method (GET, POST, etc.)
	Headers     map[string]string // HTTP headers
	Payload     interface{}       // Task payload (will be JSON encoded)
	ScheduleAt  time.Time         // When to execute the task
	RetryConfig *RetryConfig      // Retry configuration
}

// RetryConfig defines retry behavior for tasks
type RetryConfig struct {
	MaxAttempts       int           // Maximum number of attempts
	MaxRetryDuration  time.Duration // Maximum time to retry
	MinBackoff        time.Duration // Minimum backoff duration
	MaxBackoff        time.Duration // Maximum backoff duration
	MaxDoublings      int           // Maximum number of backoff doublings
}

// QueueStats contains statistics about a queue
type QueueStats struct {
	TasksCount      int64
	OldestTaskAge   time.Duration
	ExecutedLastMin int64
	ConcurrentTasks int64
}

// CloudTaskQueue implements TaskQueue using Google Cloud Tasks
type CloudTaskQueue struct {
	projectID string
	location  string
	// In production, this would use the Cloud Tasks client
	// For now, we'll simulate it
}

// LocalTaskQueue implements TaskQueue using in-memory storage for development
type LocalTaskQueue struct {
	tasks map[string][]*Task // queue -> tasks
	mu    sync.RWMutex
	
	// Background processor
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewTaskQueue creates a new task queue based on the environment
func NewTaskQueue(ctx context.Context) (TaskQueue, error) {
	if isLocalDevelopment() {
		common.Info("[TASKS] Initializing LocalTaskQueue for development")
		return NewLocalTaskQueue(), nil
	}
	
	common.Info("[TASKS] Initializing CloudTaskQueue for production")
	return NewCloudTaskQueue(ctx)
}

// NewCloudTaskQueue creates a new cloud-based task queue
func NewCloudTaskQueue(ctx context.Context) (*CloudTaskQueue, error) {
	projectID := os.Getenv("PROJECT_ID")
	if projectID == "" {
		projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}
	
	location := os.Getenv("CLOUD_TASKS_LOCATION")
	if location == "" {
		location = "us-central1"
	}
	
	if projectID == "" {
		return nil, fmt.Errorf("PROJECT_ID not configured")
	}
	
	// In production, initialize Cloud Tasks client here
	
	return &CloudTaskQueue{
		projectID: projectID,
		location:  location,
	}, nil
}

// NewLocalTaskQueue creates a new local in-memory task queue
func NewLocalTaskQueue() *LocalTaskQueue {
	q := &LocalTaskQueue{
		tasks:    make(map[string][]*Task),
		stopChan: make(chan struct{}),
	}
	
	// Start background processor
	q.wg.Add(1)
	go q.processLocalTasks()
	
	return q
}

// CreateTask creates a new task in the cloud queue
func (q *CloudTaskQueue) CreateTask(ctx context.Context, task *Task) error {
	// In production, this would use the Cloud Tasks API
	// For now, we'll simulate it
	common.Info("[TASKS] Would create cloud task: queue=%s, url=%s", task.Queue, task.URL)
	return nil
}

// CreateHTTPTask creates an HTTP task in the cloud queue
func (q *CloudTaskQueue) CreateHTTPTask(ctx context.Context, queueName string, url string, payload interface{}) error {
	task := &Task{
		Queue:   queueName,
		URL:     url,
		Method:  "POST",
		Payload: payload,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
	
	return q.CreateTask(ctx, task)
}

// DeleteTask deletes a task from the cloud queue
func (q *CloudTaskQueue) DeleteTask(ctx context.Context, queueName string, taskName string) error {
	// In production, this would use the Cloud Tasks API
	common.Info("[TASKS] Would delete cloud task: queue=%s, task=%s", queueName, taskName)
	return nil
}

// PurgeQueue purges all tasks from a cloud queue
func (q *CloudTaskQueue) PurgeQueue(ctx context.Context, queueName string) error {
	// In production, this would use the Cloud Tasks API
	common.Info("[TASKS] Would purge cloud queue: %s", queueName)
	return nil
}

// GetQueueStats returns statistics for a cloud queue
func (q *CloudTaskQueue) GetQueueStats(ctx context.Context, queueName string) (*QueueStats, error) {
	// In production, this would use the Cloud Tasks API
	return &QueueStats{
		TasksCount:      0,
		OldestTaskAge:   0,
		ExecutedLastMin: 0,
		ConcurrentTasks: 0,
	}, nil
}

// CreateTask creates a new task in the local queue
func (q *LocalTaskQueue) CreateTask(ctx context.Context, task *Task) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	
	if task.Queue == "" {
		task.Queue = "default"
	}
	
	if task.Name == "" {
		task.Name = fmt.Sprintf("task-%d", time.Now().UnixNano())
	}
	
	q.tasks[task.Queue] = append(q.tasks[task.Queue], task)
	
	common.Debug("[LOCAL_TASKS] Created task: queue=%s, name=%s, url=%s", 
		task.Queue, task.Name, task.URL)
	
	return nil
}

// CreateHTTPTask creates an HTTP task in the local queue
func (q *LocalTaskQueue) CreateHTTPTask(ctx context.Context, queueName string, url string, payload interface{}) error {
	task := &Task{
		Queue:   queueName,
		URL:     url,
		Method:  "POST",
		Payload: payload,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
	
	return q.CreateTask(ctx, task)
}

// DeleteTask deletes a task from the local queue
func (q *LocalTaskQueue) DeleteTask(ctx context.Context, queueName string, taskName string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	
	tasks, ok := q.tasks[queueName]
	if !ok {
		return fmt.Errorf("queue not found: %s", queueName)
	}
	
	for i, task := range tasks {
		if task.Name == taskName {
			q.tasks[queueName] = append(tasks[:i], tasks[i+1:]...)
			common.Debug("[LOCAL_TASKS] Deleted task: queue=%s, name=%s", queueName, taskName)
			return nil
		}
	}
	
	return fmt.Errorf("task not found: %s", taskName)
}

// PurgeQueue purges all tasks from a local queue
func (q *LocalTaskQueue) PurgeQueue(ctx context.Context, queueName string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	
	count := len(q.tasks[queueName])
	delete(q.tasks, queueName)
	
	common.Info("[LOCAL_TASKS] Purged %d tasks from queue: %s", count, queueName)
	return nil
}

// GetQueueStats returns statistics for a local queue
func (q *LocalTaskQueue) GetQueueStats(ctx context.Context, queueName string) (*QueueStats, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	
	tasks, ok := q.tasks[queueName]
	if !ok {
		return &QueueStats{}, nil
	}
	
	stats := &QueueStats{
		TasksCount: int64(len(tasks)),
	}
	
	// Find oldest task
	var oldestTime time.Time
	for _, task := range tasks {
		if oldestTime.IsZero() || task.ScheduleAt.Before(oldestTime) {
			oldestTime = task.ScheduleAt
		}
	}
	
	if !oldestTime.IsZero() {
		stats.OldestTaskAge = time.Since(oldestTime)
	}
	
	return stats, nil
}

// Stop stops the local task queue processor
func (q *LocalTaskQueue) Stop() {
	close(q.stopChan)
	q.wg.Wait()
}

// processLocalTasks processes tasks in the background (for local development)
func (q *LocalTaskQueue) processLocalTasks() {
	defer q.wg.Done()
	
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-q.stopChan:
			return
		case <-ticker.C:
			q.processPendingTasks()
		}
	}
}

func (q *LocalTaskQueue) processPendingTasks() {
	q.mu.Lock()
	defer q.mu.Unlock()
	
	now := time.Now()
	
	for queueName, tasks := range q.tasks {
		var remaining []*Task
		
		for _, task := range tasks {
			// Check if task is ready to execute
			if task.ScheduleAt.IsZero() || task.ScheduleAt.Before(now) {
				// Execute task in background
				go q.executeLocalTask(task)
			} else {
				remaining = append(remaining, task)
			}
		}
		
		q.tasks[queueName] = remaining
	}
}

func (q *LocalTaskQueue) executeLocalTask(task *Task) {
	common.Debug("[LOCAL_TASKS] Executing task: %s -> %s", task.Name, task.URL)
	
	// In local development, we just log the task execution
	// In production, this would make an actual HTTP request
	
	if task.Payload != nil {
		data, _ := json.MarshalIndent(task.Payload, "", "  ")
		common.Debug("[LOCAL_TASKS] Task payload: %s", string(data))
	}
}

// Helper functions

func isLocalDevelopment() bool {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = os.Getenv("GAE_ENV")
	}
	return env == "" || env == "development" || env == "local"
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:      10,
		MaxRetryDuration: 1 * time.Hour,
		MinBackoff:       1 * time.Second,
		MaxBackoff:       10 * time.Minute,
		MaxDoublings:     5,
	}
}

// StandardQueues provides standard queue names
var StandardQueues = struct {
	Default   string
	Critical  string
	Bulk      string
	Scheduled string
}{
	Default:   "default",
	Critical:  "critical",
	Bulk:      "bulk",
	Scheduled: "scheduled",
}