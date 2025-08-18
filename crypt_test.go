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

func TestEncryptKeyDerivation(t *testing.T) {
	// Test that encryption with short key works (now uses SHA-256 derivation)
	msg := "test message"
	shortKey := "abc"
	enc := Encrypt(context.Background(), shortKey, msg)
	if enc == "" {
		t.Fatal("encryption with short key failed")
	}
	
	// Verify ciphertext length is appropriate (nonce + ciphertext + auth tag)
	// Hex encoding doubles the length
	if len(enc) < 32 { // At least 16 bytes (12 nonce + some ciphertext) * 2 for hex
		t.Fatalf("ciphertext too short: %d bytes", len(enc))
	}
	
	// Verify decryption with same key works
	dec := Decrypt(context.Background(), shortKey, enc)
	if dec != msg {
		t.Fatalf("decryption failed: got %q, want %q", dec, msg)
	}
	
	// Verify decryption with different key fails
	wrongDec := Decrypt(context.Background(), "xyz", enc)
	if wrongDec != "" {
		t.Fatalf("decryption with wrong key should fail, got %q", wrongDec)
	}
}

func TestEncryptEmptyMessage(t *testing.T) {
	// Test that empty message can be encrypted
	enc := Encrypt(context.Background(), "key", "")
	if enc == "" {
		t.Fatal("encryption of empty message failed")
	}
	
	dec := Decrypt(context.Background(), "key", enc)
	if dec != "" {
		t.Fatalf("decryption of empty message failed: got %q", dec)
	}
}

func TestDeriveKey(t *testing.T) {
	// Test that deriveKey produces consistent 32-byte keys
	key1 := deriveKey("test")
	key2 := deriveKey("test")
	
	if len(key1) != 32 {
		t.Fatalf("derived key should be 32 bytes, got %d", len(key1))
	}
	
	// Same input should produce same key
	for i := range key1 {
		if key1[i] != key2[i] {
			t.Fatal("deriveKey not deterministic")
		}
	}
	
	// Different input should produce different key
	key3 := deriveKey("different")
	same := true
	for i := range key1 {
		if key1[i] != key3[i] {
			same = false
			break
		}
	}
	if same {
		t.Fatal("different inputs produced same derived key")
	}
}
