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
