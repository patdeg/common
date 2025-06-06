package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
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

var commonIV = []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f}

func Encrypt(c context.Context, key string, message string) string {
	// Create the aes encryption algorithm
	myKey := "yellow submarine" + key
	ciph, err := aes.NewCipher([]byte(myKey[len(myKey)-16:]))
	if err != nil {
		Error("Error NewCipher: %v", err)
		return ""
	}
	// Encrypted string
	cfb := cipher.NewCFBEncrypter(ciph, commonIV)
	ciphertext := make([]byte, len(message))
	cfb.XORKeyStream(ciphertext, []byte(message))
	return hex.EncodeToString(ciphertext)
}

func Decrypt(c context.Context, key string, message string) string {

	// Create the aes encryption algorithm
	myKey := "yellow submarine" + key
	ciph, err := aes.NewCipher([]byte(myKey[len(myKey)-16:]))
	if err != nil {
		Error("Error NewCipher: %v", err)
		return ""
	}

	messageByte, err := hex.DecodeString(message)
	if err != nil {
		Error("Error Decoding string: %v", err)
		return ""
	}

	// Decrypt strings
	cfbdec := cipher.NewCFBDecrypter(ciph, commonIV)
	plaintext := make([]byte, len(messageByte))
	cfbdec.XORKeyStream(plaintext, messageByte)
	return string(plaintext)

}
