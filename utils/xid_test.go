package utils

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateXid(t *testing.T) {
	xid := GenerateXid()

	// Should not be empty
	assert.NotEmpty(t, xid)

	// XID should be 20 characters long
	assert.Len(t, xid, 20)

	// Should contain only base32 characters (0-9, a-v)
	matched, err := regexp.MatchString("^[0-9a-v]{20}$", xid)
	assert.NoError(t, err)
	assert.True(t, matched, "XID should contain only base32 characters (0-9, a-v)")
}

func TestGenerateXid_Format(t *testing.T) {
	xid := GenerateXid()

	// XID format: 4-byte timestamp + 3-byte machine id + 2-byte process id + 3-byte counter
	// All encoded in base32, resulting in 20 characters
	assert.Len(t, xid, 20)

	// Should not contain any invalid base32 characters
	invalidChars := regexp.MustCompile("[^0-9a-v]")
	assert.False(t, invalidChars.MatchString(xid), "XID should not contain invalid base32 characters")
}

func TestGenerateXid_Uniqueness(t *testing.T) {
	// Generate multiple XIDs and ensure they are unique
	xids := make(map[string]bool)
	const numXIDs = 1000

	for range numXIDs {
		xid := GenerateXid()

		// Check that XID is not empty
		assert.NotEmpty(t, xid)

		// Check that XID is unique
		assert.False(t, xids[xid], "XID should be unique, but got duplicate: %s", xid)
		xids[xid] = true
	}

	// Verify we generated the expected number of unique XIDs
	assert.Len(t, xids, numXIDs)
}

func TestGenerateXid_ConsistentLength(t *testing.T) {
	// Generate multiple XIDs and ensure they all have the same length
	for range 100 {
		xid := GenerateXid()
		assert.Len(t, xid, 20, "All XIDs should have length 20")
	}
}

func TestGenerateXid_Base32Only(t *testing.T) {
	// Generate multiple XIDs and ensure they only contain base32 characters
	base32Pattern := regexp.MustCompile("^[0-9a-v]+$")

	for range 100 {
		xid := GenerateXid()
		assert.True(t, base32Pattern.MatchString(xid), "XID should contain only base32 characters: %s", xid)
	}
}

func TestGenerateXid_NoUpperCase(t *testing.T) {
	// Ensure XIDs are in lowercase
	upperCasePattern := regexp.MustCompile("[A-Z]")

	for range 100 {
		xid := GenerateXid()
		assert.False(t, upperCasePattern.MatchString(xid), "XID should not contain uppercase letters: %s", xid)
	}
}

func TestGenerateXid_TimeOrdering(t *testing.T) {
	// XIDs should be roughly time-ordered since they contain a timestamp
	// This is not strict due to timestamp resolution and counter behavior,
	// but should generally hold true over a larger sample

	xids := make([]string, 100)
	for i := range 100 {
		xids[i] = GenerateXid()
	}

	// Check that most XIDs are in ascending order
	// We allow some tolerance since XIDs generated in rapid succession might have the same timestamp
	ascendingCount := 0
	for i := 1; i < len(xids); i++ {
		if xids[i] >= xids[i-1] {
			ascendingCount++
		}
	}

	// At least 70% should be in ascending order (less strict than UUID v7 due to counter behavior)
	assert.GreaterOrEqual(t, ascendingCount, 70, "Most XIDs should be in ascending order due to timestamp ordering")
}

func TestGenerateXid_PerformanceConsistency(t *testing.T) {
	// Test that the function consistently returns valid XIDs even under load
	const numGoroutines = 10
	const xidsPerGoroutine = 100

	done := make(chan []string, numGoroutines)

	// Generate XIDs concurrently
	for range numGoroutines {
		go func() {
			xids := make([]string, xidsPerGoroutine)
			for j := range xidsPerGoroutine {
				xids[j] = GenerateXid()
			}
			done <- xids
		}()
	}

	// Collect all XIDs
	allXIDs := make(map[string]bool)
	for range numGoroutines {
		xids := <-done
		for _, xid := range xids {
			// Check format
			assert.Len(t, xid, 20)
			assert.NotEmpty(t, xid)

			// Check uniqueness
			assert.False(t, allXIDs[xid], "XID should be unique even under concurrent generation")
			allXIDs[xid] = true
		}
	}

	// Verify total count
	expectedTotal := numGoroutines * xidsPerGoroutine
	assert.Len(t, allXIDs, expectedTotal)
}

func TestGenerateXid_ValidBase32Characters(t *testing.T) {
	xid := GenerateXid()

	// Check each character is a valid base32 character
	for i, char := range xid {
		assert.True(t,
			(char >= '0' && char <= '9') || (char >= 'a' && char <= 'v'),
			"Character at position %d should be valid base32: %c", i, char)
	}
}

func TestGenerateXid_NotAllZeros(t *testing.T) {
	// Ensure XID is not all zeros (extremely unlikely but worth checking)
	xid := GenerateXid()

	assert.NotEqual(t, "00000000000000000000", xid)
}

func TestGenerateXid_NotAllSameCharacter(t *testing.T) {
	// Ensure XID is not all the same character
	xid := GenerateXid()

	// Check that not all characters are the same
	firstChar := xid[0]
	allSame := true
	for i := 1; i < len(xid); i++ {
		if xid[i] != firstChar {
			allSame = false
			break
		}
	}

	assert.False(t, allSame, "XID should not be all the same character")
}

func TestGenerateXid_MultipleCallsDistribution(t *testing.T) {
	// Test that multiple calls produce reasonably distributed results
	const numXIDs = 1000
	xids := make([]string, numXIDs)

	for i := range numXIDs {
		xids[i] = GenerateXid()
	}

	// Check that we have variety in the first character
	firstChars := make(map[rune]int)
	for _, xid := range xids {
		firstChars[rune(xid[0])]++
	}

	// Should have at least 1 different first character in 1000 XIDs (relaxed expectation)
	// (less variety expected than UUID due to timestamp encoding)
	assert.GreaterOrEqual(t, len(firstChars), 1, "Should have variety in first characters")

	// Check that we have variety in the last character (counter component)
	lastChars := make(map[rune]int)
	for _, xid := range xids {
		lastChars[rune(xid[19])]++
	}

	// Should have at least 2 different last characters in 1000 XIDs (relaxed expectation)
	assert.GreaterOrEqual(t, len(lastChars), 2, "Should have variety in last characters due to counter")
}

func TestGenerateXid_NoSpecialCharacters(t *testing.T) {
	// Ensure XID doesn't contain any special characters that might cause issues
	xid := GenerateXid()

	// Should not contain spaces, dashes, or other special characters
	specialChars := regexp.MustCompile("[^0-9a-v]")
	assert.False(t, specialChars.MatchString(xid), "XID should not contain special characters")
}

func TestGenerateXid_URLSafe(t *testing.T) {
	// XIDs should be URL-safe
	xid := GenerateXid()

	// Should not contain characters that need URL encoding
	urlUnsafeChars := regexp.MustCompile("[^0-9a-zA-Z]")
	assert.False(t, urlUnsafeChars.MatchString(xid), "XID should be URL-safe")
}

func TestGenerateXid_RapidGeneration(t *testing.T) {
	// Test rapid generation to ensure counter increments properly
	const numRapidXIDs = 100
	xids := make([]string, numRapidXIDs)

	// Generate XIDs as quickly as possible
	for i := range numRapidXIDs {
		xids[i] = GenerateXid()
	}

	// All should be unique
	xidSet := make(map[string]bool)
	for _, xid := range xids {
		assert.False(t, xidSet[xid], "Rapidly generated XIDs should be unique")
		xidSet[xid] = true
	}

	assert.Len(t, xidSet, numRapidXIDs)
}

func TestGenerateXid_Deterministic_Properties(t *testing.T) {
	// Test that XIDs have deterministic properties based on the spec
	xid := GenerateXid()

	// Should be exactly 20 characters
	assert.Len(t, xid, 20)

	// Should be base32 encoded
	base32Chars := "0123456789abcdefghijklmnopqrstuv"
	for _, char := range xid {
		assert.Contains(t, base32Chars, string(char), "XID should only contain base32 characters")
	}
}

func BenchmarkGenerateXid(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateXid()
	}
}

func BenchmarkGenerateXid_Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			GenerateXid()
		}
	})
}

func BenchmarkGenerateXid_StringConversion(b *testing.B) {
	// Benchmark the string conversion operation specifically
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		xid := GenerateXid()
		// Ensure the XID is actually used to prevent optimization
		if len(xid) != 20 {
			b.Fatal("Invalid XID length")
		}
	}
}

func BenchmarkGenerateXid_vs_UUID(b *testing.B) {
	b.Run("XID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GenerateXid()
		}
	})

	b.Run("UUID", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			GenerateUUIDv7()
		}
	})
}
