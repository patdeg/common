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
	"io"
	"os"

	"golang.org/x/net/context"
)

// GetContent reads the file named by filename and returns its contents.
// Any errors encountered are logged and returned.
func GetContent(c context.Context, filename string) (*[]byte, error) {
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
