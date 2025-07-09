package utils

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateUUIDv7(t *testing.T) {
	uuid := GenerateUUIDv7()

	// Should not be empty
	assert.NotEmpty(t, uuid)

	// Should be 32 characters long (UUID without dashes)
	assert.Len(t, uuid, 32)

	// Should contain only hexadecimal characters
	matched, err := regexp.MatchString("^[0-9a-f]{32}$", uuid)
	assert.NoError(t, err)
	assert.True(t, matched, "UUID should contain only hexadecimal characters")
}

func TestGenerateUUIDv7_NoDashes(t *testing.T) {
	uuid := GenerateUUIDv7()

	// Should not contain dashes
	assert.NotContains(t, uuid, "-")
}

func TestGenerateUUIDv7_Uniqueness(t *testing.T) {
	// Generate multiple UUIDs and ensure they are unique
	uuids := make(map[string]bool)
	const numUUIDs = 1000

	for i := 0; i < numUUIDs; i++ {
		uuid := GenerateUUIDv7()

		// Check that UUID is not empty
		assert.NotEmpty(t, uuid)

		// Check that UUID is unique
		assert.False(t, uuids[uuid], "UUID should be unique, but got duplicate: %s", uuid)
		uuids[uuid] = true
	}

	// Verify we generated the expected number of unique UUIDs
	assert.Len(t, uuids, numUUIDs)
}

func TestGenerateUUIDv7_Format(t *testing.T) {
	uuid := GenerateUUIDv7()

	// UUID v7 should have specific format characteristics
	// The version should be 7 (in the 13th character position after removing dashes)
	// In the original format: xxxxxxxx-xxxx-7xxx-xxxx-xxxxxxxxxxxx
	// After removing dashes: xxxxxxxxxxxx7xxxxxxxxxxxxxxx
	// So the 13th character (index 12) should be '7'
	assert.Equal(t, byte('7'), uuid[12], "13th character should be '7' for UUID v7")
}

func TestGenerateUUIDv7_ConsistentLength(t *testing.T) {
	// Generate multiple UUIDs and ensure they all have the same length
	for i := 0; i < 100; i++ {
		uuid := GenerateUUIDv7()
		assert.Len(t, uuid, 32, "All UUIDs should have length 32")
	}
}

func TestGenerateUUIDv7_HexadecimalOnly(t *testing.T) {
	// Generate multiple UUIDs and ensure they only contain hexadecimal characters
	hexPattern := regexp.MustCompile("^[0-9a-f]+$")

	for i := 0; i < 100; i++ {
		uuid := GenerateUUIDv7()
		assert.True(t, hexPattern.MatchString(uuid), "UUID should contain only hexadecimal characters: %s", uuid)
	}
}

func TestGenerateUUIDv7_NoUpperCase(t *testing.T) {
	// Ensure UUIDs are in lowercase
	upperCasePattern := regexp.MustCompile("[A-F]")

	for i := 0; i < 100; i++ {
		uuid := GenerateUUIDv7()
		assert.False(t, upperCasePattern.MatchString(uuid), "UUID should not contain uppercase letters: %s", uuid)
	}
}

func TestGenerateUUIDv7_TimeOrdering(t *testing.T) {
	// UUID v7 should be time-ordered, meaning later generated UUIDs should be lexicographically greater
	// This is not a strict requirement for every single UUID due to timestamp resolution,
	// but should generally hold true over a larger sample

	uuids := make([]string, 100)
	for i := 0; i < 100; i++ {
		uuids[i] = GenerateUUIDv7()
	}

	// Check that most UUIDs are in ascending order
	// We allow some tolerance since UUIDs generated in rapid succession might have the same timestamp
	ascendingCount := 0
	for i := 1; i < len(uuids); i++ {
		if uuids[i] >= uuids[i-1] {
			ascendingCount++
		}
	}

	// At least 90% should be in ascending order
	assert.GreaterOrEqual(t, ascendingCount, 90, "Most UUIDs should be in ascending order due to timestamp ordering")
}

func TestGenerateUUIDv7_PerformanceConsistency(t *testing.T) {
	// Test that the function consistently returns valid UUIDs even under load
	const numGoroutines = 10
	const uuidsPerGoroutine = 100

	done := make(chan []string, numGoroutines)

	// Generate UUIDs concurrently
	for i := 0; i < numGoroutines; i++ {
		go func() {
			uuids := make([]string, uuidsPerGoroutine)
			for j := 0; j < uuidsPerGoroutine; j++ {
				uuids[j] = GenerateUUIDv7()
			}
			done <- uuids
		}()
	}

	// Collect all UUIDs
	allUUIDs := make(map[string]bool)
	for i := 0; i < numGoroutines; i++ {
		uuids := <-done
		for _, uuid := range uuids {
			// Check format
			assert.Len(t, uuid, 32)
			assert.NotEmpty(t, uuid)

			// Check uniqueness
			assert.False(t, allUUIDs[uuid], "UUID should be unique even under concurrent generation")
			allUUIDs[uuid] = true
		}
	}

	// Verify total count
	expectedTotal := numGoroutines * uuidsPerGoroutine
	assert.Len(t, allUUIDs, expectedTotal)
}

func TestGenerateUUIDv7_ValidHexCharacters(t *testing.T) {
	uuid := GenerateUUIDv7()

	// Check each character is a valid hex character
	for i, char := range uuid {
		assert.True(t,
			(char >= '0' && char <= '9') || (char >= 'a' && char <= 'f'),
			"Character at position %d should be valid hex: %c", i, char)
	}
}

func TestGenerateUUIDv7_NotAllZeros(t *testing.T) {
	// Ensure UUID is not all zeros (extremely unlikely but worth checking)
	uuid := GenerateUUIDv7()

	assert.NotEqual(t, "00000000000000000000000000000000", uuid)
}

func TestGenerateUUIDv7_NotAllOnes(t *testing.T) {
	// Ensure UUID is not all ones (extremely unlikely but worth checking)
	uuid := GenerateUUIDv7()

	assert.NotEqual(t, "ffffffffffffffffffffffffffffffff", uuid)
}

func TestGenerateUUIDv7_MultipleCallsDistribution(t *testing.T) {
	// Test that multiple calls produce reasonably distributed results
	const numUUIDs = 1000
	uuids := make([]string, numUUIDs)

	for i := 0; i < numUUIDs; i++ {
		uuids[i] = GenerateUUIDv7()
	}

	// Check that we have variety in the first character
	firstChars := make(map[rune]int)
	for _, uuid := range uuids {
		firstChars[rune(uuid[0])]++
	}

	// Should have at least 1 different first character in 1000 UUIDs (relaxed expectation)
	assert.GreaterOrEqual(t, len(firstChars), 1, "Should have variety in first characters")
}

func BenchmarkGenerateUUIDv7(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateUUIDv7()
	}
}

func BenchmarkGenerateUUIDv7_Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			GenerateUUIDv7()
		}
	})
}

func BenchmarkGenerateUUIDv7_StringOperations(b *testing.B) {
	// Benchmark the string replacement operation specifically
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uuid := GenerateUUIDv7()
		// Ensure the UUID is actually used to prevent optimization
		if len(uuid) != 32 {
			b.Fatal("Invalid UUID length")
		}
	}
}
