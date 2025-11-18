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

// This file provides small cryptographic helpers used across the repository.
//
// SecureHash and GenerateSecureID provide cryptographically secure hashing and ID generation.
// Hash returns the CRC32 hash of a given string (for non-security checksums).
// Encrypt and Decrypt perform authenticated encryption using AES-GCM.
//
// DEPRECATED: MD5() is deprecated and should not be used for security purposes.
// Use SecureHash() for integrity checking or GenerateSecureID() for identifiers.

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/crc32"

	"golang.org/x/net/context"
)

// SecureHash generates a SHA-256 hash of the input string
// Use this for hashing non-secret data where integrity matters (e.g., asset versioning)
// For secret data, use HMAC or password hashing functions like bcrypt
func SecureHash(data string) string {
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

// GenerateSecureID creates a cryptographically secure random identifier
// Use this for session IDs, cookie IDs, tokens, or any security-sensitive identifier
// Returns a 64-character hex string (32 bytes of entropy / 256 bits)
func GenerateSecureID() (string, error) {
	b := make([]byte, 32) // 256 bits of entropy
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate secure ID: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func Hash(data string) uint32 {
	return crc32.ChecksumIEEE([]byte(data))
}

// deriveKey derives a 32-byte key from a secret string using SHA-256
func deriveKey(secret string) []byte {
	h := sha256.Sum256([]byte(secret))
	return h[:]
}

// Encrypt encrypts message using AES-256-GCM and returns a hex encoded nonce
// followed by ciphertext.
// Note: For backward compatibility, this function returns empty string on error
// and logs the error. New code should check for empty return value.
func Encrypt(c context.Context, key string, message string) string {
	derivedKey := deriveKey(key)
	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		Error("Error NewCipher: %v", err)
		return ""
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		Error("Error NewGCM: %v", err)
		return ""
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		Error("Error generating nonce: %v", err)
		return ""
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(message), nil)
	out := append(nonce, ciphertext...)
	return hex.EncodeToString(out)
}

// Decrypt decrypts a message encrypted with Encrypt.
// Note: For backward compatibility, this function returns empty string on error
// and logs the error. New code should check for empty return value.
func Decrypt(c context.Context, key string, message string) string {
	derivedKey := deriveKey(key)
	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		Error("Error NewCipher: %v", err)
		return ""
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		Error("Error NewGCM: %v", err)
		return ""
	}

	data, err := hex.DecodeString(message)
	if err != nil {
		Error("Error Decoding string: %v", err)
		return ""
	}
	if len(data) < gcm.NonceSize() {
		Error("Error: ciphertext too short")
		return ""
	}
	nonce := data[:gcm.NonceSize()]
	ciphertext := data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		Error("Error while decrypting: %v", err)
		return ""
	}
	return string(plaintext)
}
