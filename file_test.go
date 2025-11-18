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
	"path/filepath"
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

func TestValidatePath(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		basePath  string
		userPath  string
		wantErr   bool
		shouldLog string
	}{
		{
			name:     "valid relative path",
			basePath: tmpDir,
			userPath: "file.txt",
			wantErr:  false,
		},
		{
			name:      "directory traversal with ..",
			basePath:  tmpDir,
			userPath:  "../../../etc/passwd",
			wantErr:   true,
			shouldLog: "directory traversal",
		},
		{
			name:     "absolute path",
			basePath: tmpDir,
			userPath: "/etc/passwd",
			wantErr:  true,
		},
		{
			name:     "nested valid path",
			basePath: tmpDir,
			userPath: "subdir/file.txt",
			wantErr:  false,
		},
		{
			name:     "path with ./ component",
			basePath: tmpDir,
			userPath: "./file.txt",
			wantErr:  false,
		},
		{
			name:     "path with multiple ./ components",
			basePath: tmpDir,
			userPath: "./subdir/./file.txt",
			wantErr:  false,
		},
		{
			name:     "traversal in middle of path",
			basePath: tmpDir,
			userPath: "foo/../../../etc/passwd",
			wantErr:  true,
		},
		{
			name:     "empty path",
			basePath: tmpDir,
			userPath: "",
			wantErr:  false, // Clean path becomes "."
		},
		{
			name:     "single dot",
			basePath: tmpDir,
			userPath: ".",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidatePath(tt.basePath, tt.userPath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePath() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePath() unexpected error: %v", err)
				}

				// Verify result is within base path
				absBase, _ := filepath.Abs(tt.basePath)
				if !filepath.HasPrefix(result, absBase) {
					t.Errorf("Result path %s not within base %s", result, absBase)
				}
			}
		})
	}
}

func TestSymlinkAttack(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory outside the base
	outsideDir := filepath.Join(tmpDir, "..", "outside")
	os.MkdirAll(outsideDir, 0755)
	defer os.RemoveAll(outsideDir)

	// Create a symlink pointing outside the base directory
	symlinkPath := filepath.Join(tmpDir, "symlink")
	err := os.Symlink(outsideDir, symlinkPath)
	if err != nil {
		t.Skip("Symlink creation failed (may not be supported on this system)")
	}

	// Try to access through symlink
	result, err := ValidatePath(tmpDir, "symlink/secret.txt")
	t.Logf("tmpDir: %s", tmpDir)
	t.Logf("outsideDir: %s", outsideDir)
	t.Logf("result: %s, err: %v", result, err)
	if err == nil {
		t.Error("Symlink attack should be blocked")
	}
}

func TestValidatePathEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("Windows-style path separator", func(t *testing.T) {
		// filepath.Clean normalizes these, so they should be safe
		_, err := ValidatePath(tmpDir, "subdir\\file.txt")
		if err != nil {
			// This might fail on Unix systems, which is okay
			t.Logf("Windows-style path rejected: %v", err)
		}
	})

	t.Run("multiple slashes", func(t *testing.T) {
		result, err := ValidatePath(tmpDir, "subdir//file.txt")
		if err != nil {
			t.Errorf("Multiple slashes should be cleaned: %v", err)
		}
		if result == "" {
			t.Error("Result should not be empty")
		}
	})

	t.Run("trailing slash", func(t *testing.T) {
		result, err := ValidatePath(tmpDir, "subdir/")
		if err != nil {
			t.Errorf("Trailing slash should be valid: %v", err)
		}
		if result == "" {
			t.Error("Result should not be empty")
		}
	})
}
