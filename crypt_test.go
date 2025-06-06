package common

import (
	"context"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	msg := "hello world"
	key := "secret"
	enc := Encrypt(context.Background(), key, msg)
	if enc == "" {
		t.Fatal("empty ciphertext")
	}
	dec := Decrypt(context.Background(), key, enc)
	if dec != msg {
		t.Fatalf("Decrypt = %q, want %q", dec, msg)
	}
}

func TestDecryptWrongKey(t *testing.T) {
	msg := "secret data"
	enc := Encrypt(context.Background(), "k1", msg)
	dec := Decrypt(context.Background(), "k2", enc)
	if dec != "" {
		t.Fatalf("expected empty result with wrong key, got %q", dec)
	}
}

func TestDecryptInvalidHex(t *testing.T) {
	dec := Decrypt(context.Background(), "k", "invalid")
	if dec != "" {
		t.Fatalf("expected empty result for bad input, got %q", dec)
	}
}
