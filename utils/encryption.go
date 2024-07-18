package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"strings"
)

type DataEncryptionUtils struct {
	key []byte
}

func UseEncryption(key string) *DataEncryptionUtils {
	if key == "" {
		key = "OEASE$GOE@2024"
	}
	return &DataEncryptionUtils{key: []byte(key)}
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
	stream := cipher.NewCFBEncrypter(block, iv)
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

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)

	return cipherText, nil
}
