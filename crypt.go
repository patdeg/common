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
// MD5 and Hash return the MD5 checksum and CRC32 hash of a given string.
// Encrypt and Decrypt perform authenticated encryption using AES-GCM.

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"

	"golang.org/x/net/context"
)

func MD5(data string) string {
	h := md5.New()
	io.WriteString(h, data)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func Hash(data string) uint32 {
	return crc32.ChecksumIEEE([]byte(data))
}

// Encrypt encrypts message using AES-GCM and returns a hex encoded nonce
// followed by ciphertext.

func Encrypt(c context.Context, key string, message string) string {
	myKey := "yellow submarine" + key
	block, err := aes.NewCipher([]byte(myKey[len(myKey)-16:]))
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

func Decrypt(c context.Context, key string, message string) string {

	myKey := "yellow submarine" + key
	block, err := aes.NewCipher([]byte(myKey[len(myKey)-16:]))
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
