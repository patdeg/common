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

// Package common provides shared utilities used throughout the repository.
//
// This file contains simple helpers for working with slices. They are
// referenced by other packages when building lists of unique values or
// checking if a string is present.
package common

// AddIfNotExists appends element to list only if the element is not already
// present. It is useful when constructing slices that must contain unique
// values.
func AddIfNotExists(element string, list []string) []string {
	for _, item := range list {
		if item == element {
			return list
		}
	}
	return append(list, element)
}

// AddIfNotExistsGeneric performs the same conditional append as
// AddIfNotExists but works with a slice of empty interfaces. It is used when
// the type of the elements is not known ahead of time.
func AddIfNotExistsGeneric(element interface{}, list []interface{}) []interface{} {
	for _, item := range list {
		if item == element {
			return list
		}
	}
	return append(list, element)
}

// StringInSlice reports whether the string a exists in list.
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
