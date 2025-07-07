package utils_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.oease.dev/goe/v2/utils"
)

func TestArrContainsStr(t *testing.T) {
	tests := []struct {
		name     string
		array    []string
		str      string
		expected bool
	}{
		{
			name:     "string exists in array",
			array:    []string{"apple", "banana", "cherry"},
			str:      "banana",
			expected: true,
		},
		{
			name:     "string does not exist in array",
			array:    []string{"apple", "banana", "cherry"},
			str:      "orange",
			expected: false,
		},
		{
			name:     "empty array",
			array:    []string{},
			str:      "test",
			expected: false,
		},
		{
			name:     "empty string in array",
			array:    []string{"", "test", "value"},
			str:      "",
			expected: true,
		},
		{
			name:     "empty string not in array",
			array:    []string{"test", "value"},
			str:      "",
			expected: false,
		},
		{
			name:     "single element array - match",
			array:    []string{"single"},
			str:      "single",
			expected: true,
		},
		{
			name:     "single element array - no match",
			array:    []string{"single"},
			str:      "other",
			expected: false,
		},
		{
			name:     "case sensitive - different case",
			array:    []string{"Apple", "Banana", "Cherry"},
			str:      "apple",
			expected: false,
		},
		{
			name:     "case sensitive - exact match",
			array:    []string{"Apple", "Banana", "Cherry"},
			str:      "Apple",
			expected: true,
		},
		{
			name:     "duplicate values in array",
			array:    []string{"test", "value", "test", "other"},
			str:      "test",
			expected: true,
		},
		{
			name:     "whitespace string",
			array:    []string{"test", " ", "value"},
			str:      " ",
			expected: true,
		},
		{
			name:     "special characters",
			array:    []string{"test@example.com", "user#123", "value$"},
			str:      "user#123",
			expected: true,
		},
		{
			name:     "unicode characters",
			array:    []string{"测试", "тест", "テスト"},
			str:      "тест",
			expected: true,
		},
		{
			name:     "large array",
			array:    generateLargeStringArray(1000),
			str:      "item_500",
			expected: true,
		},
		{
			name:     "large array - not found",
			array:    generateLargeStringArray(1000),
			str:      "not_found",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ArrContainsStr(tt.array, tt.str)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to generate a large string array for testing
func generateLargeStringArray(size int) []string {
	arr := make([]string, size)
	for i := 0; i < size; i++ {
		arr[i] = fmt.Sprintf("item_%d", i)
	}
	return arr
}

func BenchmarkArrContainsStr(b *testing.B) {
	// Test with different array sizes
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			array := generateLargeStringArray(size)
			searchStr := fmt.Sprintf("item_%d", size/2) // Search for middle element

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				utils.ArrContainsStr(array, searchStr)
			}
		})
	}
}

func BenchmarkArrContainsStrWorstCase(b *testing.B) {
	// Test worst case - string not found
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("not_found_size_%d", size), func(b *testing.B) {
			array := generateLargeStringArray(size)
			searchStr := "not_found"

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				utils.ArrContainsStr(array, searchStr)
			}
		})
	}
}
