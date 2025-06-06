package common

import "testing"

func TestCamelCase(t *testing.T) {
	got := CamelCase("hello-world")
	want := "HelloWorld"
	if got != want {
		t.Errorf("CamelCase(\"hello-world\") = %q, want %q", got, want)
	}
}
