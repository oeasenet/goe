package utils

import (
	"crypto/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUseAesEncryption_DefaultKey(t *testing.T) {
	encryptor := UseAesEncryption("")

	assert.NotNil(t, encryptor)
	assert.Equal(t, "bda0b4de2fc68638dcac98a6603f6d2f", string(encryptor.key))
}

func TestUseAesEncryption_CustomKey(t *testing.T) {
	customKey := "my-custom-32-byte-key-for-aes!"
	encryptor := UseAesEncryption(customKey)

	assert.NotNil(t, encryptor)
	assert.Equal(t, customKey, string(encryptor.key))
}

func TestDataEncryptionUtils_EncryptDecrypt_Success(t *testing.T) {
	tests := []struct {
		name string
		data string
		key  string
	}{
		{
			name: "short string with default key",
			data: "hello",
			key:  "",
		},
		{
			name: "empty string with default key",
			data: "",
			key:  "",
		},
		{
			name: "long string with default key",
			data: strings.Repeat("test data ", 1000),
			key:  "",
		},
		{
			name: "unicode string with default key",
			data: "Hello, 世界! 🌍",
			key:  "",
		},
		{
			name: "special characters with default key",
			data: "!@#$%^&*()_+-=[]{}|;:,.<>?",
			key:  "",
		},
		{
			name: "short string with custom key",
			data: "hello",
			key:  "my-custom-32-byte-key-for-aes!!!",
		},
		{
			name: "multiline string with custom key",
			data: "Line 1\nLine 2\nLine 3\n",
			key:  "another-32-byte-key-for-testing!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encryptor := UseAesEncryption(tt.key)

			// Encrypt
			encrypted, err := encryptor.Encrypt([]byte(tt.data))
			require.NoError(t, err)
			assert.NotEmpty(t, encrypted)
			assert.NotEqual(t, tt.data, encrypted)

			// Decrypt
			decrypted, err := encryptor.Decrypt(encrypted)
			require.NoError(t, err)
			assert.Equal(t, tt.data, string(decrypted))
		})
	}
}

func TestDataEncryptionUtils_EncryptDecrypt_BinaryData(t *testing.T) {
	// Test with binary data
	binaryData := make([]byte, 256)
	for i := range binaryData {
		binaryData[i] = byte(i)
	}

	encryptor := UseAesEncryption("")

	encrypted, err := encryptor.Encrypt(binaryData)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := encryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, binaryData, decrypted)
}

func TestDataEncryptionUtils_Encrypt_RandomIV(t *testing.T) {
	// Test that each encryption produces different ciphertext due to random IV
	encryptor := UseAesEncryption("")
	data := []byte("test data")

	encrypted1, err := encryptor.Encrypt(data)
	require.NoError(t, err)

	encrypted2, err := encryptor.Encrypt(data)
	require.NoError(t, err)

	// Should be different due to random IV
	assert.NotEqual(t, encrypted1, encrypted2)

	// But both should decrypt to the same data
	decrypted1, err := encryptor.Decrypt(encrypted1)
	require.NoError(t, err)
	assert.Equal(t, data, decrypted1)

	decrypted2, err := encryptor.Decrypt(encrypted2)
	require.NoError(t, err)
	assert.Equal(t, data, decrypted2)
}

func TestDataEncryptionUtils_Encrypt_InvalidKey(t *testing.T) {
	// Test with invalid key length
	invalidKeys := []string{
		"short", // too short
		"this-key-is-too-long-for-aes-256-encryption", // too long
		"exactly-31-bytes-long-key!",                  // 31 bytes (should be 32 for AES-256)
	}

	for _, key := range invalidKeys {
		t.Run("key length "+string(rune(len(key))), func(t *testing.T) {
			encryptor := UseAesEncryption(key)

			_, err := encryptor.Encrypt([]byte("test"))
			assert.Error(t, err)
		})
	}
}

func TestDataEncryptionUtils_Decrypt_InvalidData(t *testing.T) {
	encryptor := UseAesEncryption("")

	tests := []struct {
		name        string
		invalidData string
	}{
		{
			name:        "empty string",
			invalidData: "",
		},
		{
			name:        "invalid base64",
			invalidData: "invalid-base64-data!@#",
		},
		{
			name:        "too short ciphertext",
			invalidData: "YWJjZA", // base64 for "abcd" - too short for AES block
		},
		{
			name:        "random string",
			invalidData: "c2hvcnRfaW52YWxpZA", // base64 for "short_invalid" - too short
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := encryptor.Decrypt(tt.invalidData)
			assert.Error(t, err)
		})
	}
}

func TestDataEncryptionUtils_Decrypt_Base64Padding(t *testing.T) {
	encryptor := UseAesEncryption("")
	data := []byte("test data for padding")

	encrypted, err := encryptor.Encrypt(data)
	require.NoError(t, err)

	// The encrypt function removes padding, so decrypt should handle it
	decrypted, err := encryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, data, decrypted)
}

func TestDataEncryptionUtils_Decrypt_WrongKey(t *testing.T) {
	// Encrypt with one key, try to decrypt with another
	encryptor1 := UseAesEncryption("key1-32-bytes-long-for-aes-256!!")
	encryptor2 := UseAesEncryption("key2-32-bytes-long-for-aes-256!!")

	data := []byte("sensitive data")

	encrypted, err := encryptor1.Encrypt(data)
	require.NoError(t, err)

	// Try to decrypt with different key - should succeed but return garbage
	decrypted, err := encryptor2.Decrypt(encrypted)
	require.NoError(t, err)
	assert.NotEqual(t, data, decrypted)
}

func TestDataEncryptionUtils_LargeData(t *testing.T) {
	encryptor := UseAesEncryption("")

	// Test with large data (1MB)
	largeData := make([]byte, 1024*1024)
	_, err := rand.Read(largeData)
	require.NoError(t, err)

	encrypted, err := encryptor.Encrypt(largeData)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := encryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, largeData, decrypted)
}

func TestDataEncryptionUtils_ConcurrentUsage(t *testing.T) {
	encryptor := UseAesEncryption("")

	const numGoroutines = 10
	const iterations = 100

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < iterations; j++ {
				data := []byte(strings.Repeat("test", id+1))

				encrypted, err := encryptor.Encrypt(data)
				assert.NoError(t, err)
				assert.NotEmpty(t, encrypted)

				decrypted, err := encryptor.Decrypt(encrypted)
				assert.NoError(t, err)
				assert.Equal(t, data, decrypted)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestDataEncryptionUtils_KeyLength32Bytes(t *testing.T) {
	// Test that 32-byte keys work correctly (AES-256)
	key32 := "this-is-exactly-32-bytes-long!!!"
	assert.Equal(t, 32, len(key32))

	encryptor := UseAesEncryption(key32)
	data := []byte("test data")

	encrypted, err := encryptor.Encrypt(data)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := encryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, data, decrypted)
}

func TestDataEncryptionUtils_KeyLength16Bytes(t *testing.T) {
	// Test that 16-byte keys work correctly (AES-128)
	key16 := "16-byte-key-here"
	assert.Equal(t, 16, len(key16))

	encryptor := UseAesEncryption(key16)
	data := []byte("test data")

	encrypted, err := encryptor.Encrypt(data)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := encryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, data, decrypted)
}

func TestDataEncryptionUtils_KeyLength24Bytes(t *testing.T) {
	// Test that 24-byte keys work correctly (AES-192)
	key24 := "24-byte-key-here-exactly"
	assert.Equal(t, 24, len(key24))

	encryptor := UseAesEncryption(key24)
	data := []byte("test data")

	encrypted, err := encryptor.Encrypt(data)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := encryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, data, decrypted)
}

func BenchmarkDataEncryptionUtils_Encrypt(b *testing.B) {
	encryptor := UseAesEncryption("")
	data := []byte("benchmark test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encryptor.Encrypt(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDataEncryptionUtils_Decrypt(b *testing.B) {
	encryptor := UseAesEncryption("")
	data := []byte("benchmark test data")

	encrypted, err := encryptor.Encrypt(data)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encryptor.Decrypt(encrypted)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDataEncryptionUtils_EncryptDecrypt(b *testing.B) {
	encryptor := UseAesEncryption("")
	data := []byte("benchmark test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encrypted, err := encryptor.Encrypt(data)
		if err != nil {
			b.Fatal(err)
		}

		_, err = encryptor.Decrypt(encrypted)
		if err != nil {
			b.Fatal(err)
		}
	}
}
