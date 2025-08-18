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

package api

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestDoWithRetryBytesReader(t *testing.T) {
	// Track request attempts and body content
	var attempts int32
	expectedBody := "test request body"
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		
		// Read and verify request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
		}
		
		if string(body) != expectedBody {
			t.Errorf("Request body = %q; want %q", string(body), expectedBody)
		}
		
		// Return 500 on first attempt to trigger retry
		if atomic.LoadInt32(&attempts) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()
	
	client := NewClient(ClientConfig{
		BaseURL: server.URL,
		RetryConfig: &RetryConfig{
			MaxRetries:  1,
			InitialWait: 10 * time.Millisecond,
			MaxWait:     100 * time.Millisecond,
			Multiplier:  2.0,
			RetryOn:     []int{500},
		},
	})
	
	// Create request with bytes.Reader body
	req, err := http.NewRequest("POST", server.URL+"/test", bytes.NewReader([]byte(expectedBody)))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	
	// Execute request with retry
	resp, err := client.doWithRetry(context.Background(), req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	
	// Verify response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Response status = %d; want %d", resp.StatusCode, http.StatusOK)
	}
	
	if string(resp.Body) != "success" {
		t.Errorf("Response body = %q; want %q", string(resp.Body), "success")
	}
	
	// Verify that the request was retried
	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("Request attempts = %d; want 2", atomic.LoadInt32(&attempts))
	}
}

func TestDoWithRetryNonSeekableBody(t *testing.T) {
	// Track request attempts and body content
	var attempts int32
	expectedBody := "test request body from pipe"
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		
		// Read and verify request body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
		}
		
		if string(body) != expectedBody {
			t.Errorf("Request body on attempt %d = %q; want %q", 
				atomic.LoadInt32(&attempts), string(body), expectedBody)
		}
		
		// Return 500 on first attempt to trigger retry
		if atomic.LoadInt32(&attempts) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()
	
	client := NewClient(ClientConfig{
		BaseURL: server.URL,
		RetryConfig: &RetryConfig{
			MaxRetries:  1,
			InitialWait: 10 * time.Millisecond,
			MaxWait:     100 * time.Millisecond,
			Multiplier:  2.0,
			RetryOn:     []int{500},
		},
	})
	
	// Create a pipe to simulate non-seekable body
	pr, pw := io.Pipe()
	go func() {
		pw.Write([]byte(expectedBody))
		pw.Close()
	}()
	
	// Create request with non-seekable body
	req, err := http.NewRequest("POST", server.URL+"/test", pr)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	
	// Execute request with retry
	resp, err := client.doWithRetry(context.Background(), req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	
	// Verify response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Response status = %d; want %d", resp.StatusCode, http.StatusOK)
	}
	
	if string(resp.Body) != "success" {
		t.Errorf("Response body = %q; want %q", string(resp.Body), "success")
	}
	
	// Verify that the request was retried
	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("Request attempts = %d; want 2", atomic.LoadInt32(&attempts))
	}
}

func TestDoWithRetryNoBody(t *testing.T) {
	var attempts int32
	
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		
		// Return 500 on first attempt to trigger retry
		if atomic.LoadInt32(&attempts) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()
	
	client := NewClient(ClientConfig{
		BaseURL: server.URL,
		RetryConfig: &RetryConfig{
			MaxRetries:  1,
			InitialWait: 10 * time.Millisecond,
			MaxWait:     100 * time.Millisecond,
			Multiplier:  2.0,
			RetryOn:     []int{500},
		},
	})
	
	// Create request without body
	req, err := http.NewRequest("GET", server.URL+"/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	
	// Execute request with retry
	resp, err := client.doWithRetry(context.Background(), req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	
	// Verify response
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Response status = %d; want %d", resp.StatusCode, http.StatusOK)
	}
	
	// Verify that the request was retried
	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("Request attempts = %d; want 2", atomic.LoadInt32(&attempts))
	}
}