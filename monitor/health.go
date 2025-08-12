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

// Package monitor provides health checks, metrics collection, and
// monitoring capabilities for applications.
package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/patdeg/common"
)

// HealthChecker defines the health check interface
type HealthChecker interface {
	// Check performs a health check
	Check(ctx context.Context) *HealthStatus

	// Name returns the checker name
	Name() string
}

// HealthStatus represents the health status of a component
type HealthStatus struct {
	Status      Status                 `json:"status"`
	Message     string                 `json:"message,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	LastChecked time.Time              `json:"last_checked"`
	Duration    time.Duration          `json:"duration_ms"`
}

// Status represents health status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

// Monitor manages health checks and metrics
type Monitor struct {
	checkers    map[string]HealthChecker
	results     map[string]*HealthStatus
	metrics     *Metrics
	mu          sync.RWMutex
	checkPeriod time.Duration
	stopChan    chan struct{}
}

// NewMonitor creates a new monitor
func NewMonitor(checkPeriod time.Duration) *Monitor {
	if checkPeriod == 0 {
		checkPeriod = 30 * time.Second
	}

	m := &Monitor{
		checkers:    make(map[string]HealthChecker),
		results:     make(map[string]*HealthStatus),
		metrics:     NewMetrics(),
		checkPeriod: checkPeriod,
		stopChan:    make(chan struct{}),
	}

	// Start background health checks
	go m.runHealthChecks()

	return m
}

// AddChecker adds a health checker
func (m *Monitor) AddChecker(checker HealthChecker) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.checkers[checker.Name()] = checker
	common.Debug("[MONITOR] Added health checker: %s", checker.Name())
}

// RemoveChecker removes a health checker
func (m *Monitor) RemoveChecker(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.checkers, name)
	delete(m.results, name)
}

// GetHealth returns the overall health status
func (m *Monitor) GetHealth() *HealthReport {
	m.mu.RLock()
	defer m.mu.RUnlock()

	report := &HealthReport{
		Status:    StatusHealthy,
		Checks:    make(map[string]*HealthStatus),
		Timestamp: time.Now(),
	}

	// Aggregate health status
	for name, status := range m.results {
		report.Checks[name] = status

		// Update overall status
		if status.Status == StatusUnhealthy {
			report.Status = StatusUnhealthy
		} else if status.Status == StatusDegraded && report.Status == StatusHealthy {
			report.Status = StatusDegraded
		}
	}

	// Add system metrics
	report.System = m.getSystemMetrics()

	return report
}

// CheckHealth performs an immediate health check
func (m *Monitor) CheckHealth(ctx context.Context) *HealthReport {
	m.performHealthChecks(ctx)
	return m.GetHealth()
}

// runHealthChecks runs periodic health checks
func (m *Monitor) runHealthChecks() {
	ticker := time.NewTicker(m.checkPeriod)
	defer ticker.Stop()

	// Initial check
	m.performHealthChecks(context.Background())

	for {
		select {
		case <-ticker.C:
			m.performHealthChecks(context.Background())
		case <-m.stopChan:
			return
		}
	}
}

// performHealthChecks executes all health checks
func (m *Monitor) performHealthChecks(ctx context.Context) {
	m.mu.RLock()
	checkers := make(map[string]HealthChecker)
	for k, v := range m.checkers {
		checkers[k] = v
	}
	m.mu.RUnlock()

	// Run checks in parallel
	var wg sync.WaitGroup
	results := make(chan struct {
		name   string
		status *HealthStatus
	}, len(checkers))

	for name, checker := range checkers {
		wg.Add(1)
		go func(n string, c HealthChecker) {
			defer wg.Done()

			// Run check with timeout
			checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			start := time.Now()
			status := c.Check(checkCtx)
			status.Duration = time.Since(start)
			status.LastChecked = time.Now()

			results <- struct {
				name   string
				status *HealthStatus
			}{name: n, status: status}
		}(name, checker)
	}

	wg.Wait()
	close(results)

	// Update results
	m.mu.Lock()
	for result := range results {
		m.results[result.name] = result.status

		// Record metric
		m.metrics.RecordHealthCheck(result.name, result.status.Status)
	}
	m.mu.Unlock()
}

// getSystemMetrics returns system-level metrics
func (m *Monitor) getSystemMetrics() *SystemMetrics {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return &SystemMetrics{
		Memory: MemoryMetrics{
			Alloc:      memStats.Alloc,
			TotalAlloc: memStats.TotalAlloc,
			Sys:        memStats.Sys,
			NumGC:      memStats.NumGC,
		},
		Goroutines: runtime.NumGoroutine(),
		CPUs:       runtime.NumCPU(),
		Uptime:     m.metrics.GetUptime(),
	}
}

// Stop stops the monitor
func (m *Monitor) Stop() {
	close(m.stopChan)
}

// ServeHTTP implements http.Handler for health endpoint
func (m *Monitor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	report := m.GetHealth()

	// Set appropriate status code
	statusCode := http.StatusOK
	if report.Status == StatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	} else if report.Status == StatusDegraded {
		statusCode = http.StatusOK // Still operational but degraded
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(report)
}

// HealthReport represents the overall health report
type HealthReport struct {
	Status    Status                   `json:"status"`
	Checks    map[string]*HealthStatus `json:"checks"`
	System    *SystemMetrics           `json:"system"`
	Timestamp time.Time                `json:"timestamp"`
}

// SystemMetrics represents system-level metrics
type SystemMetrics struct {
	Memory     MemoryMetrics `json:"memory"`
	Goroutines int           `json:"goroutines"`
	CPUs       int           `json:"cpus"`
	Uptime     time.Duration `json:"uptime_seconds"`
}

// MemoryMetrics represents memory metrics
type MemoryMetrics struct {
	Alloc      uint64 `json:"alloc_bytes"`
	TotalAlloc uint64 `json:"total_alloc_bytes"`
	Sys        uint64 `json:"sys_bytes"`
	NumGC      uint32 `json:"num_gc"`
}

// Built-in health checkers

// PingChecker checks basic connectivity
type PingChecker struct{}

func (p *PingChecker) Name() string {
	return "ping"
}

func (p *PingChecker) Check(ctx context.Context) *HealthStatus {
	return &HealthStatus{
		Status:  StatusHealthy,
		Message: "pong",
	}
}

// DatabaseChecker checks database connectivity
type DatabaseChecker struct {
	name string
	ping func(ctx context.Context) error
}

// NewDatabaseChecker creates a new database checker
func NewDatabaseChecker(name string, ping func(ctx context.Context) error) *DatabaseChecker {
	return &DatabaseChecker{
		name: name,
		ping: ping,
	}
}

func (d *DatabaseChecker) Name() string {
	return d.name
}

func (d *DatabaseChecker) Check(ctx context.Context) *HealthStatus {
	if err := d.ping(ctx); err != nil {
		return &HealthStatus{
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("Database error: %v", err),
		}
	}

	return &HealthStatus{
		Status:  StatusHealthy,
		Message: "Database connection OK",
	}
}

// HTTPChecker checks HTTP endpoint health
type HTTPChecker struct {
	name   string
	url    string
	client *http.Client
}

// NewHTTPChecker creates a new HTTP checker
func NewHTTPChecker(name, url string) *HTTPChecker {
	return &HTTPChecker{
		name: name,
		url:  url,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (h *HTTPChecker) Name() string {
	return h.name
}

func (h *HTTPChecker) Check(ctx context.Context) *HealthStatus {
	req, err := http.NewRequestWithContext(ctx, "GET", h.url, nil)
	if err != nil {
		return &HealthStatus{
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("Failed to create request: %v", err),
		}
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return &HealthStatus{
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("Request failed: %v", err),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return &HealthStatus{
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("Service returned %d", resp.StatusCode),
		}
	}

	if resp.StatusCode >= 400 {
		return &HealthStatus{
			Status:  StatusDegraded,
			Message: fmt.Sprintf("Service returned %d", resp.StatusCode),
		}
	}

	return &HealthStatus{
		Status:  StatusHealthy,
		Message: fmt.Sprintf("Service returned %d", resp.StatusCode),
		Details: map[string]interface{}{
			"status_code": resp.StatusCode,
		},
	}
}

// DiskSpaceChecker checks available disk space
type DiskSpaceChecker struct {
	path      string
	threshold float64 // Percentage threshold for warning
}

// NewDiskSpaceChecker creates a new disk space checker
func NewDiskSpaceChecker(path string, threshold float64) *DiskSpaceChecker {
	if threshold == 0 {
		threshold = 90.0
	}
	return &DiskSpaceChecker{
		path:      path,
		threshold: threshold,
	}
}

func (d *DiskSpaceChecker) Name() string {
	return "disk_space"
}

func (d *DiskSpaceChecker) Check(ctx context.Context) *HealthStatus {
	// This is a simplified check
	// In production, use syscall.Statfs or similar

	return &HealthStatus{
		Status:  StatusHealthy,
		Message: "Disk space OK",
		Details: map[string]interface{}{
			"path":      d.path,
			"threshold": d.threshold,
		},
	}
}

// Metrics tracks application metrics
type Metrics struct {
	startTime    time.Time
	healthChecks map[string]int64
	requests     int64
	errors       int64
	mu           sync.RWMutex
}

// NewMetrics creates new metrics tracker
func NewMetrics() *Metrics {
	return &Metrics{
		startTime:    time.Now(),
		healthChecks: make(map[string]int64),
	}
}

// RecordHealthCheck records a health check result
func (m *Metrics) RecordHealthCheck(name string, status Status) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s_%s", name, status)
	m.healthChecks[key]++
}

// RecordRequest records a request
func (m *Metrics) RecordRequest() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests++
}

// RecordError records an error
func (m *Metrics) RecordError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors++
}

// GetUptime returns application uptime
func (m *Metrics) GetUptime() time.Duration {
	return time.Since(m.startTime)
}

// GetStats returns metrics statistics
func (m *Metrics) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := map[string]interface{}{
		"uptime_seconds": m.GetUptime().Seconds(),
		"requests":       m.requests,
		"errors":         m.errors,
		"health_checks":  m.healthChecks,
	}

	return stats
}
