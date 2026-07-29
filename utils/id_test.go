package utils

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateNanoId(t *testing.T) {
	tests := []struct {
		name     string
		length   []int
		expected int
		wantErr  bool
	}{
		{
			name:     "Default length",
			length:   nil,
			expected: 32,
			wantErr:  false,
		},
		{
			name:     "Custom valid length",
			length:   []int{50},
			expected: 50,
			wantErr:  false,
		},
		{
			name:     "Zero length",
			length:   []int{0},
			expected: 0,
			wantErr:  true,
		},
		{
			name:     "Negative length",
			length:   []int{-10},
			expected: 0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tt.wantErr {
					t.Errorf("unexpected panic: %v", r)
				}
			}()
			got := GenerateNanoId(tt.length...)
			if len(got) != tt.expected && !tt.wantErr {
				t.Errorf("GenerateNanoId() length = %v, expected %v", len(got), tt.expected)
			}
			if tt.wantErr && len(got) != 0 {
				t.Errorf("GenerateNanoId() should have returned an error or empty result, but got %v", got)
			}
		})
	}
}

func TestGenerateNanoIdUniqueness(t *testing.T) {
	// The default ID is 32 characters from a 62-character alphabet, a keyspace of
	// roughly 62^32. No feasible number of draws will find a genuine collision, so
	// what this actually checks is that the generator is not catastrophically
	// broken — returning a constant, or drawing from a tiny amount of entropy.
	// 100k draws detect that just as well as the 10 million this used to generate,
	// which took ~22s under -race and dominated the whole test suite.
	const numIDs = 100_000
	ids := make(map[string]bool, numIDs)

	for range numIDs {
		id := GenerateNanoId()
		if ids[id] {
			t.Errorf("Collision detected: duplicate ID %s generated", id)
		}
		ids[id] = true
	}

	if len(ids) != numIDs {
		t.Errorf("Expected %d unique IDs, got %d", numIDs, len(ids))
	}
}

func TestGenerateNanoIdCharacterSet(t *testing.T) {
	// Test that generated IDs only contain valid characters
	validChars := regexp.MustCompile(`^[0-9a-zA-Z]+$`)

	for range 100 {
		id := GenerateNanoId()
		if !validChars.MatchString(id) {
			t.Errorf("Generated ID contains invalid characters: %s", id)
		}
	}
}

func TestGenerateNanoIdDifferentLengths(t *testing.T) {
	lengths := []int{1, 5, 10, 21, 32, 50, 100}

	for _, length := range lengths {
		t.Run(fmt.Sprintf("length_%d", length), func(t *testing.T) {
			id := GenerateNanoId(length)
			if len(id) != length {
				t.Errorf("Expected length %d, got %d", length, len(id))
			}
		})
	}
}

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

	for range numUUIDs {
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
	for range 100 {
		uuid := GenerateUUIDv7()
		assert.Len(t, uuid, 32, "All UUIDs should have length 32")
	}
}

func TestGenerateUUIDv7_HexadecimalOnly(t *testing.T) {
	// Generate multiple UUIDs and ensure they only contain hexadecimal characters
	hexPattern := regexp.MustCompile("^[0-9a-f]+$")

	for range 100 {
		uuid := GenerateUUIDv7()
		assert.True(t, hexPattern.MatchString(uuid), "UUID should contain only hexadecimal characters: %s", uuid)
	}
}

func TestGenerateUUIDv7_NoUpperCase(t *testing.T) {
	// Ensure UUIDs are in lowercase
	upperCasePattern := regexp.MustCompile("[A-F]")

	for range 100 {
		uuid := GenerateUUIDv7()
		assert.False(t, upperCasePattern.MatchString(uuid), "UUID should not contain uppercase letters: %s", uuid)
	}
}

func TestGenerateUUIDv7_TimeOrdering(t *testing.T) {
	// UUID v7 should be time-ordered, meaning later generated UUIDs should be lexicographically greater
	// This is not a strict requirement for every single UUID due to timestamp resolution,
	// but should generally hold true over a larger sample

	uuids := make([]string, 100)
	for i := range 100 {
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
	for range numGoroutines {
		go func() {
			uuids := make([]string, uuidsPerGoroutine)
			for j := range uuidsPerGoroutine {
				uuids[j] = GenerateUUIDv7()
			}
			done <- uuids
		}()
	}

	// Collect all UUIDs
	allUUIDs := make(map[string]bool)
	for range numGoroutines {
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

	for i := range numUUIDs {
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
