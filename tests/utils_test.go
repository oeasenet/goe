package tests

import (
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
		{"bytes", "100", 100},
		{"kilobytes", "1K", 1000},
		{"megabytes", "1M", 1000000},
		{"gigabytes", "1G", 1000000000},
		{"terabytes", "1T", 1000000000000},
		{"petabytes", "1P", 1000000000000000},
		{"decimal", "1.5K", 1500},
		{"lowercase", "1k", 1000},
		{"mixed", "2.5M", 2500000},
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
		{"zero_bytes", 0, "0B"},
		{"single_byte", 1, "1.00B"},
		{"bytes_under_1000", 999, "999.00B"},
		{"exactly_1_kilobyte", 1000, "1.00KB"},
		{"kilobytes_with_decimal", 1500, "1.50KB"},
		{"large_kilobytes", 999000, "999.00KB"},
		{"exactly_1_megabyte", 1000000, "1.00MB"},
		{"megabytes_with_decimal", 1500000, "1.50MB"},
		{"large_megabytes", 999000000, "999.00MB"},
		{"exactly_1_gigabyte", 1000000000, "1.00GB"},
		{"gigabytes_with_decimal", 1500000000, "1.50GB"},
		{"large_gigabytes", 999000000000, "999.00GB"},
		{"exactly_1_terabyte", 1000000000000, "1.00TB"},
		{"terabytes_with_decimal", 1500000000000, "1.50TB"},
		{"exactly_1_petabyte", 1000000000000000, "1.00PB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ConvertBytesToHumanReadableSize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRoundTripConversion(t *testing.T) {
	tests := []struct {
		name  string
		input int
	}{
		{"roundtrip_1000", 1000},
		{"roundtrip_1500", 1500},
		{"roundtrip_1000000", 1000000},
		{"roundtrip_2500000", 2500000},
		{"roundtrip_1000000000", 1000000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			humanReadable := utils.ConvertBytesToHumanReadableSize(tt.input)
			backToBytes := utils.HumanReadableSizeToBytes(humanReadable)
			assert.Equal(t, tt.input, backToBytes)
		})
	}
}

func TestArrContainsStr(t *testing.T) {
	tests := []struct {
		name     string
		arr      []string
		str      string
		expected bool
	}{
		{"string_exists_in_array", []string{"apple", "banana", "cherry"}, "banana", true},
		{"string_does_not_exist_in_array", []string{"apple", "banana", "cherry"}, "grape", false},
		{"empty_array", []string{}, "test", false},
		{"empty_string_in_array", []string{"", "test"}, "", true},
		{"empty_string_not_in_array", []string{"test", "array"}, "", false},
		{"single_element_array_-_match", []string{"only"}, "only", true},
		{"single_element_array_-_no_match", []string{"only"}, "different", false},
		{"case_sensitive_-_different_case", []string{"Apple", "Banana"}, "apple", false},
		{"case_sensitive_-_exact_match", []string{"Apple", "Banana"}, "Apple", true},
		{"duplicate_values_in_array", []string{"test", "test", "value"}, "test", true},
		{"whitespace_string", []string{"  ", "test"}, "  ", true},
		{"special_characters", []string{"@#$", "test"}, "@#$", true},
		{"unicode_characters", []string{"🙂", "test"}, "🙂", true},
		{"large_array", make([]string, 1000), "not_found", false},
		{"large_array_-_not_found", append(make([]string, 1000), "found"), "found", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ArrContainsStr(tt.arr, tt.str)
			assert.Equal(t, tt.expected, result)
		})
	}
}
