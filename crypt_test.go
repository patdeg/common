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
