package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strings"
)

// DataEncryptionUtils provides AES encryption and decryption utilities.
type DataEncryptionUtils struct {
	key []byte
}

// UseAesEncryption creates a new AES encryption utility with the given key.
// The key must be 16, 24, or 32 bytes long for AES-128, AES-192, or AES-256 respectively.
// An empty key is not permitted.
func UseAesEncryption(key string) (*DataEncryptionUtils, error) {
	if key == "" {
		return nil, errors.New("encryption key must not be empty")
	}
	keyLen := len(key)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return nil, fmt.Errorf("encryption key must be 16, 24, or 32 bytes for AES-128/192/256, got %d", keyLen)
	}
	return &DataEncryptionUtils{key: []byte(key)}, nil
}

func (deu *DataEncryptionUtils) Encrypt(data []byte) (string, error) {
	block, err := aes.NewCipher(deu.key)
	if err != nil {
		return "", err
	}

	cipherText := make([]byte, aes.BlockSize+len(data))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	stream := cipher.NewCFBEncrypter(block, iv) //nolint:staticcheck // CFB kept for backward compatibility with existing encrypted data
	stream.XORKeyStream(cipherText[aes.BlockSize:], data)
	return strings.TrimRight(base64.URLEncoding.EncodeToString(cipherText), "="), nil
}

func (deu *DataEncryptionUtils) Decrypt(data string) ([]byte, error) {
	// Calculate the number of padding characters needed
	switch len(data) % 4 {
	case 2:
		data += "=="
	case 3:
		data += "="
	}
	cipherText, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(deu.key)
	if err != nil {
		return nil, err
	}

	if len(cipherText) < aes.BlockSize {
		return nil, errors.New("invalid cipher text block size")
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv) //nolint:staticcheck // CFB kept for backward compatibility with existing encrypted data
	stream.XORKeyStream(cipherText, cipherText)

	return cipherText, nil
}
