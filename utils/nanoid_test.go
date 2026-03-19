package utils

import (
	"fmt"
	"regexp"
	"testing"
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
	// Test collision resistance by generating many IDs
	const numIDs = 10000000
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
