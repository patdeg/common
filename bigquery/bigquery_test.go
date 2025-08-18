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

package bigquery

import (
	"errors"
	"testing"

	"google.golang.org/api/googleapi"
)

func TestIsTableNotFoundError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "googleapi 404 error",
			err:  &googleapi.Error{Code: 404},
			want: true,
		},
		{
			name: "googleapi other error",
			err:  &googleapi.Error{Code: 500},
			want: false,
		},
		{
			name: "error with 'not found' message",
			err:  errors.New("table not found in dataset"),
			want: true,
		},
		{
			name: "error with 'does not exist' message",
			err:  errors.New("table does not exist"),
			want: true,
		},
		{
			name: "unrelated error",
			err:  errors.New("connection timeout"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTableNotFoundError(tt.err); got != tt.want {
				t.Errorf("isTableNotFoundError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestIsAlreadyExistsError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "googleapi 409 error",
			err:  &googleapi.Error{Code: 409},
			want: true,
		},
		{
			name: "googleapi other error",
			err:  &googleapi.Error{Code: 404},
			want: false,
		},
		{
			name: "error with 'already exists' message",
			err:  errors.New("table already exists"),
			want: true,
		},
		{
			name: "unrelated error",
			err:  errors.New("permission denied"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAlreadyExistsError(tt.err); got != tt.want {
				t.Errorf("isAlreadyExistsError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// Benchmark to verify the new implementation is faster
func BenchmarkStringContains(b *testing.B) {
	longString := "This is a very long string that we're searching for a substring within to test performance"
	searchString := "substring"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isTableNotFoundError(errors.New(longString + " not found " + searchString))
	}
}