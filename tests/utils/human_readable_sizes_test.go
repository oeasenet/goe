package utils_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.oease.dev/goe/v2/utils"
)

func TestHumanReadableSizeToBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		// Basic byte values
		{
			name:     "bytes only",
			input:    "100",
			expected: 100,
		},
		{
			name:     "zero bytes",
			input:    "0",
			expected: 0,
		},

		// Kilobytes
		{
			name:     "kilobytes lowercase",
			input:    "1k",
			expected: 1000,
		},
		{
			name:     "kilobytes uppercase",
			input:    "1K",
			expected: 1000,
		},
		{
			name:     "kilobytes with decimal",
			input:    "1.5k",
			expected: 1500,
		},
		{
			name:     "kilobytes with space",
			input:    "2 k",
			expected: 2000,
		},

		// Megabytes
		{
			name:     "megabytes lowercase",
			input:    "1m",
			expected: 1000000,
		},
		{
			name:     "megabytes uppercase",
			input:    "1M",
			expected: 1000000,
		},
		{
			name:     "megabytes with decimal",
			input:    "2.5m",
			expected: 2500000,
		},

		// Gigabytes
		{
			name:     "gigabytes lowercase",
			input:    "1g",
			expected: 1000000000,
		},
		{
			name:     "gigabytes uppercase",
			input:    "1G",
			expected: 1000000000,
		},
		{
			name:     "gigabytes with decimal",
			input:    "1.5g",
			expected: 1500000000,
		},

		// Terabytes
		{
			name:     "terabytes lowercase",
			input:    "1t",
			expected: 1000000000000,
		},
		{
			name:     "terabytes uppercase",
			input:    "1T",
			expected: 1000000000000,
		},

		// Petabytes
		{
			name:     "petabytes lowercase",
			input:    "1p",
			expected: 1000000000000000,
		},
		{
			name:     "petabytes uppercase",
			input:    "1P",
			expected: 1000000000000000,
		},

		// Edge cases
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "invalid format",
			input:    "abc",
			expected: 0,
		},
		{
			name:     "invalid unit",
			input:    "100x",
			expected: 100, // Should ignore invalid unit and return just the number
		},
		{
			name:     "multiple spaces",
			input:    "100  k",
			expected: 100000,
		},
		{
			name:     "trailing spaces",
			input:    "100k   ",
			expected: 100000,
		},
		{
			name:     "large number",
			input:    "999999",
			expected: 999999,
		},
		{
			name:     "fractional bytes",
			input:    "100.5",
			expected: 100, // Should truncate to int
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.HumanReadableSizeToBytes(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertBytesToHumanReadableSize(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		// Basic cases
		{
			name:     "zero bytes",
			input:    0,
			expected: "0B",
		},
		{
			name:     "single byte",
			input:    1,
			expected: "1.00B",
		},
		{
			name:     "bytes under 1000",
			input:    999,
			expected: "999.00B",
		},

		// Kilobytes
		{
			name:     "exactly 1 kilobyte",
			input:    1000,
			expected: "1.00KB",
		},
		{
			name:     "kilobytes with decimal",
			input:    1500,
			expected: "1.50KB",
		},
		{
			name:     "large kilobytes",
			input:    999000,
			expected: "999.00KB",
		},

		// Megabytes
		{
			name:     "exactly 1 megabyte",
			input:    1000000,
			expected: "1.00MB",
		},
		{
			name:     "megabytes with decimal",
			input:    2500000,
			expected: "2.50MB",
		},
		{
			name:     "large megabytes",
			input:    999000000,
			expected: "999.00MB",
		},

		// Gigabytes
		{
			name:     "exactly 1 gigabyte",
			input:    1000000000,
			expected: "1.00GB",
		},
		{
			name:     "gigabytes with decimal",
			input:    1500000000,
			expected: "1.50GB",
		},
		{
			name:     "large gigabytes",
			input:    999000000000,
			expected: "999.00GB",
		},

		// Terabytes
		{
			name:     "exactly 1 terabyte",
			input:    1000000000000,
			expected: "1.00TB",
		},
		{
			name:     "terabytes with decimal",
			input:    2500000000000,
			expected: "2.50TB",
		},

		// Petabytes
		{
			name:     "exactly 1 petabyte",
			input:    1000000000000000,
			expected: "1.00PB",
		},

		// Edge cases
		{
			name:     "small fractional kilobyte",
			input:    1001,
			expected: "1.00KB", // Should round to 2 decimal places
		},
		{
			name:     "complex decimal",
			input:    1234567,
			expected: "1.23MB",
		},
		{
			name:     "very large number",
			input:    1234567890123456,
			expected: "1.23PB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ConvertBytesToHumanReadableSize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRoundTripConversion(t *testing.T) {
	// Test that converting to human readable and back gives approximately the same result
	testCases := []int{
		1000,       // 1KB
		1500,       // 1.5KB
		1000000,    // 1MB
		2500000,    // 2.5MB
		1000000000, // 1GB
	}

	for _, original := range testCases {
		t.Run(fmt.Sprintf("roundtrip_%d", original), func(t *testing.T) {
			// Convert to human readable
			humanReadable := utils.ConvertBytesToHumanReadableSize(original)

			// Convert back to bytes
			converted := utils.HumanReadableSizeToBytes(humanReadable)

			// Should be approximately equal (within 1% due to rounding)
			tolerance := float64(original) * 0.01
			diff := float64(converted - original)
			if diff < 0 {
				diff = -diff
			}

			assert.True(t, diff <= tolerance,
				"Round trip conversion failed: %d -> %s -> %d (diff: %.0f, tolerance: %.0f)",
				original, humanReadable, converted, diff, tolerance)
		})
	}
}

func BenchmarkHumanReadableSizeToBytes(b *testing.B) {
	testCases := []string{
		"100",
		"1k",
		"1.5m",
		"2g",
		"1t",
	}

	for _, tc := range testCases {
		b.Run(tc, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				utils.HumanReadableSizeToBytes(tc)
			}
		})
	}
}

func BenchmarkConvertBytesToHumanReadableSize(b *testing.B) {
	testCases := []int{
		100,
		1000,
		1000000,
		1000000000,
		1000000000000,
	}

	for _, tc := range testCases {
		b.Run(fmt.Sprintf("%d", tc), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				utils.ConvertBytesToHumanReadableSize(tc)
			}
		})
	}
}
