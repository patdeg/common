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
