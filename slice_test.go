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

import "testing"

func TestAddIfNotExists(t *testing.T) {
	list := []string{"a", "b"}
	list = AddIfNotExists("c", list)
	if len(list) != 3 || list[2] != "c" {
		t.Errorf("AddIfNotExists failed to append new element: %v", list)
	}
	list = AddIfNotExists("b", list)
	if len(list) != 3 {
		t.Errorf("AddIfNotExists modified slice when element already existed: %v", list)
	}
}

func TestAddIfNotExistsGeneric(t *testing.T) {
	list := []interface{}{1, "two"}
	list = AddIfNotExistsGeneric(3, list)
	if len(list) != 3 || list[2] != 3 {
		t.Errorf("AddIfNotExistsGeneric failed to append new element: %v", list)
	}
	list = AddIfNotExistsGeneric("two", list)
	if len(list) != 3 {
		t.Errorf("AddIfNotExistsGeneric modified slice when element already existed: %v", list)
	}
}

func TestStringInSlice(t *testing.T) {
	list := []string{"a", "b"}
	if !StringInSlice("a", list) {
		t.Error("StringInSlice did not find existing element")
	}
	if StringInSlice("c", list) {
		t.Error("StringInSlice found non-existent element")
	}
}
