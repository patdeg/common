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

package common

import (
	"bytes"
	"context"
	"os"
	"testing"
)

func TestGetContentSmallFile(t *testing.T) {
	f, err := os.CreateTemp("", "small")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString("hello"); err != nil {
		t.Fatal(err)
	}
	f.Close()

	b, err := GetContent(context.Background(), f.Name())
	if err != nil {
		t.Fatalf("GetContent returned error: %v", err)
	}
	if string(*b) != "hello" {
		t.Errorf("got %q want %q", string(*b), "hello")
	}
}

func TestGetContentLargeFile(t *testing.T) {
	f, err := os.CreateTemp("", "large")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	data := bytes.Repeat([]byte("a"), 12*1024*1024)
	if _, err := f.Write(data); err != nil {
		t.Fatal(err)
	}
	f.Close()

	b, err := GetContent(context.Background(), f.Name())
	if err != nil {
		t.Fatalf("GetContent returned error: %v", err)
	}
	if len(*b) != len(data) {
		t.Errorf("len=%d want %d", len(*b), len(data))
	}
}
