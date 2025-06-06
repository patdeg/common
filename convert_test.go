// Package common contains tests for conversion helpers.
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
