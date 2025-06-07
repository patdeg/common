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

// Package common contains small helper routines used across different
// packages. The tests in this file verify the behaviour of the conversion
// utilities implemented in convert.go.
package common

import "testing"

// TestCamelCase verifies that CamelCase converts dash separated strings into
// camel cased words.

func TestCamelCase(t *testing.T) {
	got := CamelCase("hello-world")
	want := "HelloWorld"
	if got != want {
		t.Errorf("CamelCase(\"hello-world\") = %q, want %q", got, want)
	}
}

// TestToString verifies that values of different basic types are converted
// to their string representation as expected.
func TestToString(t *testing.T) {
	tests := []struct {
		name string
		in   interface{}
		want string
	}{
		{"nil", nil, ""},
		{"int", 5, "5"},
		{"int64", int64(9), "9"},
		{"float64", 1.23, "1.23000000"},
		{"string", "foo", "foo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToString(tt.in); got != tt.want {
				t.Errorf("ToString(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestS2F ensures that S2F parses numeric strings and returns zero for
// malformed values.

func TestS2F(t *testing.T) {
	tests := []struct {
		in   string
		want float64
	}{
		{"42", 42},
		{"3.14", 3.14},
		{"bogus", 0},
	}
	for _, tt := range tests {
		if got := S2F(tt.in); got != tt.want {
			t.Errorf("S2F(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

// TestRound exercises rounding with various precisions and negative numbers.
func TestRound(t *testing.T) {
	tests := []struct {
		name      string
		n         float64
		precision int
		want      float64
	}{
		{"up", 1.235, 2, 1.24},
		{"down", 1.234, 2, 1.23},
		{"negative", -1.235, 2, -1.24},
		{"zero", 1.5, 0, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Round(tt.n, tt.precision); got != tt.want {
				t.Errorf("Round(%v,%d) = %v, want %v", tt.n, tt.precision, got, tt.want)
			}
		})
	}
}
