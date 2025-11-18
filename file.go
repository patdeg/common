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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/context"
)

// ValidatePath validates that a user-provided path is safe and within basePath
// Returns the absolute, validated path or an error if validation fails
// This function prevents path traversal attacks by:
//  1. Cleaning the path (removes .., resolves ./, etc.)
//  2. Rejecting paths containing .. after cleaning
//  3. Rejecting absolute paths in user input
//  4. Ensuring the final path is within the base directory
//  5. Handling symlinks securely
func ValidatePath(basePath, userPath string) (string, error) {
	// Clean the user path (removes .., resolves ./, etc.)
	cleanPath := filepath.Clean(userPath)

	// Reject any path containing .. after cleaning
	if strings.Contains(cleanPath, "..") {
		return "", fmt.Errorf("invalid path: contains directory traversal sequence")
	}

	// Reject absolute paths in user input
	if filepath.IsAbs(cleanPath) {
		return "", fmt.Errorf("invalid path: absolute paths not allowed")
	}

	// Get absolute base path
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base path: %w", err)
	}

	// Join base and user path
	fullPath := filepath.Join(absBase, cleanPath)

	// Evaluate symlinks in the path to detect symlink attacks
	// We check each component of the path
	checkPath := fullPath
	for {
		realPath, err := filepath.EvalSymlinks(checkPath)
		if err == nil {
			// Path or partial path exists, check if it's within base
			absReal, err := filepath.Abs(realPath)
			if err != nil {
				return "", fmt.Errorf("failed to resolve symlink: %w", err)
			}

			relPath, err := filepath.Rel(absBase, absReal)
			if err != nil || strings.HasPrefix(relPath, "..") {
				return "", fmt.Errorf("invalid path: symlink target outside base directory")
			}
			break
		}

		// Path doesn't exist, try parent directory
		parent := filepath.Dir(checkPath)
		if parent == checkPath || parent == "." || parent == "/" {
			// Reached root without finding existing path, allow it
			break
		}
		checkPath = parent
	}

	// Final check: ensure the full path would be within base
	relPath, err := filepath.Rel(absBase, fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to determine relative path: %w", err)
	}

	if strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("invalid path: outside base directory")
	}

	return fullPath, nil
}

// GetContent reads the file named by filename and returns its contents.
// Any errors encountered are logged and returned.
//
// SECURITY NOTE: This function does NOT validate paths. If accepting user input,
// use ValidatePath() first to prevent path traversal attacks.
func GetContent(c context.Context, filename string) (*[]byte, error) {
	// #nosec G304 -- callers must validate filename (e.g., with ValidatePath) before calling.
	file, err := os.Open(filename)
	if err != nil {
		Error("Error opening file %s: %v", filename, err)
		return nil, err
	}
	defer file.Close()

	Info("FILE FOUND : %s", filename)
	content, err := io.ReadAll(file)
	if err != nil {
		Error("Error reading file %s: %v", filename, err)
		return nil, err
	}

	return &content, nil

}
