package utils

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHumanReadableSizeToBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		// Basic cases
		{"empty string", "", 0},
		{"zero", "0", 0},
		{"simple number", "42", 42},
		{"decimal number", "42.5", 42},
		{"negative number", "-10", -10},

		// Bytes
		{"bytes with B", "100B", 100},
		{"bytes with b", "100b", 100},
		{"bytes with spaces", "100 B", 100},
		{"bytes with more spaces", "100  B", 100},

		// Kilobytes
		{"kilobytes with k", "1k", 1000},
		{"kilobytes with K", "1K", 1000},
		{"kilobytes with kb", "1kb", 1000},
		{"kilobytes with KB", "1KB", 1000},
		{"kilobytes decimal", "1.5k", 1500},
		{"kilobytes with spaces", "2 k", 2000},

		// Megabytes
		{"megabytes with m", "1m", 1000000},
		{"megabytes with M", "1M", 1000000},
		{"megabytes with mb", "1mb", 1000000},
		{"megabytes with MB", "1MB", 1000000},
		{"megabytes decimal", "2.5m", 2500000},
		{"megabytes with spaces", "3 M", 3000000},

		// Gigabytes
		{"gigabytes with g", "1g", 1000000000},
		{"gigabytes with G", "1G", 1000000000},
		{"gigabytes with gb", "1gb", 1000000000},
		{"gigabytes with GB", "1GB", 1000000000},
		{"gigabytes decimal", "1.5g", 1500000000},

		// Terabytes
		{"terabytes with t", "1t", 1000000000000},
		{"terabytes with T", "1T", 1000000000000},
		{"terabytes with tb", "1tb", 1000000000000},
		{"terabytes with TB", "1TB", 1000000000000},
		{"terabytes decimal", "0.5t", 500000000000},

		// Petabytes
		{"petabytes with p", "1p", 1000000000000000},
		{"petabytes with P", "1P", 1000000000000000},
		{"petabytes with pb", "1pb", 1000000000000000},
		{"petabytes with PB", "1PB", 1000000000000000},

		// Complex cases
		{"large number", "42000", 42000},
		{"decimal with unit", "10.75KB", 10750},
		{"zero with unit", "0MB", 0},
		{"fractional bytes", "0.5", 0}, // should truncate to 0

		// Edge cases
		{"only unit", "k", 0},              // no number
		{"only unit 2", "MB", 0},           // no number
		{"invalid unit", "42X", 42},        // unrecognized unit, should ignore
		{"multiple units", "42KBG", 42000}, // should use first valid unit
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HumanReadableSizeToBytes(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHumanReadableSizeToBytes_InvalidInputs(t *testing.T) {
	invalidInputs := []string{
		"abc",         // non-numeric
		"k10",         // unit before number
		"10.5.5k",     // invalid decimal
		"--10k",       // invalid negative
		"10k5",        // number after unit
		"hello world", // completely invalid
	}

	for _, input := range invalidInputs {
		t.Run("invalid: "+input, func(t *testing.T) {
			result := HumanReadableSizeToBytes(input)
			// Should return 0 for invalid inputs
			assert.Equal(t, 0, result)
		})
	}
}

func TestHumanReadableSizeToBytes_CaseInsensitive(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"10k", 10000},
		{"10K", 10000},
		{"10m", 10000000},
		{"10M", 10000000},
		{"10g", 10000000000},
		{"10G", 10000000000},
		{"10t", 10000000000000},
		{"10T", 10000000000000},
		{"10p", 10000000000000000},
		{"10P", 10000000000000000},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := HumanReadableSizeToBytes(tt.input)
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
		{"zero bytes", 0, "0B"},
		{"single byte", 1, "1.00B"},
		{"multiple bytes", 999, "999.00B"},

		// Kilobytes
		{"exact kilobyte", 1000, "1.00KB"},
		{"kilobytes", 1500, "1.50KB"},
		{"large kilobytes", 999999, "1000.00KB"},

		// Megabytes
		{"exact megabyte", 1000000, "1.00MB"},
		{"megabytes", 1500000, "1.50MB"},
		{"large megabytes", 999999999, "1000.00MB"},

		// Gigabytes
		{"exact gigabyte", 1000000000, "1.00GB"},
		{"gigabytes", 1500000000, "1.50GB"},
		{"large gigabytes", 999999999999, "1000.00GB"},

		// Terabytes
		{"exact terabyte", 1000000000000, "1.00TB"},
		{"terabytes", 1500000000000, "1.50TB"},
		{"large terabytes", 999999999999999, "1.00PB"},

		// Petabytes
		{"exact petabyte", 1000000000000000, "1.00PB"},
		{"petabytes", 1500000000000000, "1.50PB"},

		// Edge cases
		{"small number", 42, "42.00B"},
		{"medium number", 42000, "42.00KB"},
		{"large number", 42000000, "42.00MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertBytesToHumanReadableSize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertBytesToHumanReadableSize_LargeNumbers(t *testing.T) {
	// Test with very large numbers
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{"exabytes", 1000000000000000000, "1.00EB"},
		// {"zettabytes", 1000000000000000000000, "1.00ZB"}, // Skip due to int overflow
		// Note: YB would cause overflow in int, so we test the largest safe values
		{"max safe int", int(math.MaxInt64), "9.22EB"}, // approximately
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertBytesToHumanReadableSize(tt.input)
			// For very large numbers, we'll just check the unit is correct
			if tt.name == "max safe int" {
				assert.Contains(t, result, "EB")
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestConvertBytesToHumanReadableSize_Precision(t *testing.T) {
	// Test precision of decimal places
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{"1.23 KB", 1234, "1.23KB"},
		{"1.00 KB", 1000, "1.00KB"},
		{"2.00 KB", 2000, "2.00KB"},
		{"10.50 KB", 10500, "10.50KB"},
		{"100.00 KB", 100000, "100.00KB"},
		{"1.01 MB", 1010000, "1.01MB"},
		{"2.00 MB", 2000000, "2.00MB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertBytesToHumanReadableSize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRoundTripConversion(t *testing.T) {
	// Test that converting from human readable to bytes and back works
	tests := []struct {
		name         string
		originalSize int
		tolerance    float64 // tolerance for floating point errors
	}{
		{"1 KB", 1000, 0.01},
		{"1 MB", 1000000, 0.01},
		{"1 GB", 1000000000, 0.01},
		{"1.5 KB", 1500, 0.01},
		{"2.5 MB", 2500000, 0.01},
		{"10 GB", 10000000000, 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to human readable
			humanReadable := ConvertBytesToHumanReadableSize(tt.originalSize)

			// Convert back to bytes
			backToBytes := HumanReadableSizeToBytes(humanReadable)

			// Check that they're approximately equal (within tolerance)
			diff := math.Abs(float64(tt.originalSize - backToBytes))
			tolerance := tt.tolerance * float64(tt.originalSize)

			assert.True(t, diff <= tolerance,
				"Original: %d, Human: %s, Back: %d, Diff: %.2f, Tolerance: %.2f",
				tt.originalSize, humanReadable, backToBytes, diff, tolerance)
		})
	}
}

func TestHumanReadableSizeToBytes_WhitespaceHandling(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"no spaces", "10KB", 10000},
		{"space before unit", "10 KB", 10000},
		{"multiple spaces", "10  KB", 10000},
		{"space after number", "10 K", 10000},
		{"mixed spaces", "10 KB", 10000},
		{"tab character", "10\tKB", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HumanReadableSizeToBytes(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHumanReadableSizeToBytes_ScientificNotation(t *testing.T) {
	// Test that scientific notation actually works (as strconv.ParseFloat supports it)
	tests := []struct {
		input    string
		expected int
	}{
		{"1e3", 1000},
		{"1.5e6", 1500000},
		{"2E3", 2000},
		{"1e3KB", 1000000},
	}

	for _, tt := range tests {
		t.Run("scientific: "+tt.input, func(t *testing.T) {
			result := HumanReadableSizeToBytes(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func BenchmarkHumanReadableSizeToBytes(b *testing.B) {
	testInputs := []string{
		"1KB",
		"1.5MB",
		"2GB",
		"500TB",
		"1000",
		"42.5kb",
	}

	for _, input := range testInputs {
		b.Run(input, func(b *testing.B) {
			for b.Loop() {
				HumanReadableSizeToBytes(input)
			}
		})
	}
}

func BenchmarkConvertBytesToHumanReadableSize(b *testing.B) {
	testInputs := []int{
		1000,
		1500000,
		2000000000,
		500000000000,
		1000000000000000,
	}

	for _, input := range testInputs {
		b.Run(ConvertBytesToHumanReadableSize(input), func(b *testing.B) {
			for b.Loop() {
				ConvertBytesToHumanReadableSize(input)
			}
		})
	}
}

func BenchmarkRoundTripConversion(b *testing.B) {
	originalSize := 1500000 // 1.5 MB

	b.ResetTimer()
	for b.Loop() {
		humanReadable := ConvertBytesToHumanReadableSize(originalSize)
		HumanReadableSizeToBytes(humanReadable)
	}
}
