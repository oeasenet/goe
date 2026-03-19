package utils

import (
	"crypto/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testDefaultKey is a 32-byte key used as default in tests.
const testDefaultKey = "bda0b4de2fc68638dcac98a6603f6d2f"

func TestUseAesEncryption_EmptyKey(t *testing.T) {
	_, err := UseAesEncryption("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "encryption key must not be empty")
}

func TestUseAesEncryption_InvalidKeyLength(t *testing.T) {
	invalidKeys := []struct {
		name string
		key  string
	}{
		{"5 bytes", "short"},
		{"too long", "this-key-is-too-long-for-aes-256-encryption!!"},
		{"31 bytes", "exactly-31-bytes-long-key!!!!!"},
	}

	for _, tt := range invalidKeys {
		t.Run(tt.name, func(t *testing.T) {
			_, err := UseAesEncryption(tt.key)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "encryption key must be 16, 24, or 32 bytes")
		})
	}
}

func TestUseAesEncryption_ValidKeyLengths(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"AES-128 (16 bytes)", "16-byte-key-here"},
		{"AES-192 (24 bytes)", "24-byte-key-here-exactly"},
		{"AES-256 (32 bytes)", "this-is-exactly-32-bytes-long!!!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encryptor, err := UseAesEncryption(tt.key)
			require.NoError(t, err)
			assert.NotNil(t, encryptor)
			assert.Equal(t, tt.key, string(encryptor.key))
		})
	}
}

func TestDataEncryptionUtils_EncryptDecrypt_Success(t *testing.T) {
	tests := []struct {
		name string
		data string
		key  string
	}{
		{
			name: "short string",
			data: "hello",
			key:  testDefaultKey,
		},
		{
			name: "empty string",
			data: "",
			key:  testDefaultKey,
		},
		{
			name: "long string",
			data: strings.Repeat("test data ", 1000),
			key:  testDefaultKey,
		},
		{
			name: "unicode string",
			data: "Hello, 世界! 🌍",
			key:  testDefaultKey,
		},
		{
			name: "special characters",
			data: "!@#$%^&*()_+-=[]{}|;:,.<>?",
			key:  testDefaultKey,
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
			encryptor, err := UseAesEncryption(tt.key)
			require.NoError(t, err)

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
	binaryData := make([]byte, 256)
	for i := range binaryData {
		binaryData[i] = byte(i)
	}

	encryptor, err := UseAesEncryption(testDefaultKey)
	require.NoError(t, err)

	encrypted, err := encryptor.Encrypt(binaryData)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := encryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, binaryData, decrypted)
}

func TestDataEncryptionUtils_Encrypt_RandomIV(t *testing.T) {
	encryptor, err := UseAesEncryption(testDefaultKey)
	require.NoError(t, err)
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

func TestDataEncryptionUtils_Decrypt_InvalidData(t *testing.T) {
	encryptor, err := UseAesEncryption(testDefaultKey)
	require.NoError(t, err)

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
	encryptor, err := UseAesEncryption(testDefaultKey)
	require.NoError(t, err)
	data := []byte("test data for padding")

	encrypted, err := encryptor.Encrypt(data)
	require.NoError(t, err)

	decrypted, err := encryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, data, decrypted)
}

func TestDataEncryptionUtils_Decrypt_WrongKey(t *testing.T) {
	encryptor1, err := UseAesEncryption("key1-32-bytes-long-for-aes-256!!")
	require.NoError(t, err)
	encryptor2, err := UseAesEncryption("key2-32-bytes-long-for-aes-256!!")
	require.NoError(t, err)

	data := []byte("sensitive data")

	encrypted, err := encryptor1.Encrypt(data)
	require.NoError(t, err)

	// Try to decrypt with different key - should succeed but return garbage
	decrypted, err := encryptor2.Decrypt(encrypted)
	require.NoError(t, err)
	assert.NotEqual(t, data, decrypted)
}

func TestDataEncryptionUtils_LargeData(t *testing.T) {
	encryptor, err := UseAesEncryption(testDefaultKey)
	require.NoError(t, err)

	// Test with large data (1MB)
	largeData := make([]byte, 1024*1024)
	_, err = rand.Read(largeData)
	require.NoError(t, err)

	encrypted, err := encryptor.Encrypt(largeData)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := encryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, largeData, decrypted)
}

func TestDataEncryptionUtils_ConcurrentUsage(t *testing.T) {
	encryptor, err := UseAesEncryption(testDefaultKey)
	require.NoError(t, err)

	const numGoroutines = 10
	const iterations = 100

	done := make(chan bool, numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			for range iterations {
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

	for range numGoroutines {
		<-done
	}
}

func TestDataEncryptionUtils_KeyLength32Bytes(t *testing.T) {
	key32 := "this-is-exactly-32-bytes-long!!!"
	assert.Equal(t, 32, len(key32))

	encryptor, err := UseAesEncryption(key32)
	require.NoError(t, err)
	data := []byte("test data")

	encrypted, err := encryptor.Encrypt(data)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := encryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, data, decrypted)
}

func TestDataEncryptionUtils_KeyLength16Bytes(t *testing.T) {
	key16 := "16-byte-key-here"
	assert.Equal(t, 16, len(key16))

	encryptor, err := UseAesEncryption(key16)
	require.NoError(t, err)
	data := []byte("test data")

	encrypted, err := encryptor.Encrypt(data)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := encryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, data, decrypted)
}

func TestDataEncryptionUtils_KeyLength24Bytes(t *testing.T) {
	key24 := "24-byte-key-here-exactly"
	assert.Equal(t, 24, len(key24))

	encryptor, err := UseAesEncryption(key24)
	require.NoError(t, err)
	data := []byte("test data")

	encrypted, err := encryptor.Encrypt(data)
	require.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	decrypted, err := encryptor.Decrypt(encrypted)
	require.NoError(t, err)
	assert.Equal(t, data, decrypted)
}

func BenchmarkDataEncryptionUtils_Encrypt(b *testing.B) {
	encryptor, err := UseAesEncryption(testDefaultKey)
	if err != nil {
		b.Fatal(err)
	}
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
	encryptor, err := UseAesEncryption(testDefaultKey)
	if err != nil {
		b.Fatal(err)
	}
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
	encryptor, err := UseAesEncryption(testDefaultKey)
	if err != nil {
		b.Fatal(err)
	}
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
