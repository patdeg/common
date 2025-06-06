package common

import "testing"

func TestCamelCase(t *testing.T) {
	got := CamelCase("hello-world")
	want := "HelloWorld"
	if got != want {
		t.Errorf("CamelCase(\"hello-world\") = %q, want %q", got, want)
	}
}

func TestToString(t *testing.T) {
	var nilVal interface{}
	cases := []struct {
		in   interface{}
		want string
	}{
		{123, "123"},
		{int64(42), "42"},
		{3.14, "3.14000000"},
		{"foo", "foo"},
		{true, "true"},
		{nilVal, ""},
	}
	for _, c := range cases {
		if got := ToString(c.in); got != c.want {
			t.Errorf("ToString(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestS2F(t *testing.T) {
	tests := []struct {
		in   string
		want float64
	}{
		{"1.5", 1.5},
		{"1", 1},
		{"foo", 0},
	}
	for _, tt := range tests {
		if got := S2F(tt.in); got != tt.want {
			t.Errorf("S2F(%q) = %f, want %f", tt.in, got, tt.want)
		}
	}
}

func TestRound(t *testing.T) {
	cases := []struct {
		in        float64
		precision int
		want      float64
	}{
		{3.14159, 2, 3.14},
		{2.675, 2, 2.68},
		{-1.2345, 2, -1.23},
	}
	for _, c := range cases {
		if got := Round(c.in, c.precision); got != c.want {
			t.Errorf("Round(%f, %d) = %f, want %f", c.in, c.precision, got, c.want)
		}
	}
}
