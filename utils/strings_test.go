package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArrContainsStr_Found(t *testing.T) {
	tests := []struct {
		name   string
		array  []string
		str    string
		expect bool
	}{
		{
			name:   "found in small array",
			array:  []string{"apple", "banana", "cherry"},
			str:    "banana",
			expect: true,
		},
		{
			name:   "found at beginning",
			array:  []string{"first", "second", "third"},
			str:    "first",
			expect: true,
		},
		{
			name:   "found at end",
			array:  []string{"first", "second", "third"},
			str:    "third",
			expect: true,
		},
		{
			name:   "found in single element array",
			array:  []string{"only"},
			str:    "only",
			expect: true,
		},
		{
			name:   "found empty string",
			array:  []string{"", "test", ""},
			str:    "",
			expect: true,
		},
		{
			name:   "found in large array",
			array:  []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "target", "k", "l", "m", "n"},
			str:    "target",
			expect: true,
		},
		{
			name:   "found duplicate string",
			array:  []string{"test", "test", "test"},
			str:    "test",
			expect: true,
		},
		{
			name:   "found with special characters",
			array:  []string{"hello", "world!", "@#$%", "123"},
			str:    "@#$%",
			expect: true,
		},
		{
			name:   "found with spaces",
			array:  []string{"hello world", "test string", "another test"},
			str:    "test string",
			expect: true,
		},
		{
			name:   "found unicode string",
			array:  []string{"hello", "世界", "🌍", "test"},
			str:    "世界",
			expect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ArrContainsStr(tt.array, tt.str)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestArrContainsStr_NotFound(t *testing.T) {
	tests := []struct {
		name   string
		array  []string
		str    string
		expect bool
	}{
		{
			name:   "not found in small array",
			array:  []string{"apple", "banana", "cherry"},
			str:    "orange",
			expect: false,
		},
		{
			name:   "not found - case sensitive",
			array:  []string{"Apple", "Banana", "Cherry"},
			str:    "apple",
			expect: false,
		},
		{
			name:   "not found - partial match",
			array:  []string{"testing", "test123", "tested"},
			str:    "test",
			expect: false,
		},
		{
			name:   "not found - empty string in non-empty array",
			array:  []string{"test", "hello", "world"},
			str:    "",
			expect: false,
		},
		{
			name:   "not found - whitespace differences",
			array:  []string{"hello world", "test string"},
			str:    "hello  world",
			expect: false,
		},
		{
			name:   "not found - similar but different",
			array:  []string{"test1", "test2", "test3"},
			str:    "test",
			expect: false,
		},
		{
			name:   "not found in large array",
			array:  []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n"},
			str:    "target",
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ArrContainsStr(tt.array, tt.str)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestArrContainsStr_EmptyArray(t *testing.T) {
	tests := []struct {
		name   string
		array  []string
		str    string
		expect bool
	}{
		{
			name:   "empty array with non-empty string",
			array:  []string{},
			str:    "test",
			expect: false,
		},
		{
			name:   "empty array with empty string",
			array:  []string{},
			str:    "",
			expect: false,
		},
		{
			name:   "nil array with string",
			array:  nil,
			str:    "test",
			expect: false,
		},
		{
			name:   "nil array with empty string",
			array:  nil,
			str:    "",
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ArrContainsStr(tt.array, tt.str)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestArrContainsStr_ArrayWithEmptyStrings(t *testing.T) {
	tests := []struct {
		name   string
		array  []string
		str    string
		expect bool
	}{
		{
			name:   "array with empty strings - find empty",
			array:  []string{"", "test", ""},
			str:    "",
			expect: true,
		},
		{
			name:   "array with empty strings - find non-empty",
			array:  []string{"", "test", ""},
			str:    "test",
			expect: true,
		},
		{
			name:   "array with empty strings - not found",
			array:  []string{"", "test", ""},
			str:    "notfound",
			expect: false,
		},
		{
			name:   "array with only empty strings",
			array:  []string{"", "", ""},
			str:    "",
			expect: true,
		},
		{
			name:   "array with only empty strings - not found",
			array:  []string{"", "", ""},
			str:    "test",
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ArrContainsStr(tt.array, tt.str)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestArrContainsStr_DuplicateStrings(t *testing.T) {
	tests := []struct {
		name   string
		array  []string
		str    string
		expect bool
	}{
		{
			name:   "array with duplicates - found",
			array:  []string{"test", "test", "test"},
			str:    "test",
			expect: true,
		},
		{
			name:   "array with duplicates - not found",
			array:  []string{"test", "test", "test"},
			str:    "notfound",
			expect: false,
		},
		{
			name:   "mixed array with duplicates",
			array:  []string{"a", "b", "a", "c", "b", "d"},
			str:    "b",
			expect: true,
		},
		{
			name:   "mixed array with duplicates - not found",
			array:  []string{"a", "b", "a", "c", "b", "d"},
			str:    "e",
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ArrContainsStr(tt.array, tt.str)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestArrContainsStr_LargeArray(t *testing.T) {
	// Test with a large array to ensure performance is reasonable
	largeArray := make([]string, 10000)
	for i := 0; i < 10000; i++ {
		largeArray[i] = "item" + string(rune(i))
	}

	// Add target at the end
	largeArray[9999] = "target"

	tests := []struct {
		name   string
		array  []string
		str    string
		expect bool
	}{
		{
			name:   "large array - found at end",
			array:  largeArray,
			str:    "target",
			expect: true,
		},
		{
			name:   "large array - found at beginning",
			array:  largeArray,
			str:    "item0",
			expect: true,
		},
		{
			name:   "large array - not found",
			array:  largeArray,
			str:    "notfound",
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ArrContainsStr(tt.array, tt.str)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestArrContainsStr_SpecialCharacters(t *testing.T) {
	tests := []struct {
		name   string
		array  []string
		str    string
		expect bool
	}{
		{
			name:   "special characters - found",
			array:  []string{"!@#$%", "^&*()", "{}[]", "|\\:;"},
			str:    "^&*()",
			expect: true,
		},
		{
			name:   "special characters - not found",
			array:  []string{"!@#$%", "^&*()", "{}[]", "|\\:;"},
			str:    "~`",
			expect: false,
		},
		{
			name:   "newline characters",
			array:  []string{"line1\nline2", "test\n", "\n"},
			str:    "test\n",
			expect: true,
		},
		{
			name:   "tab characters",
			array:  []string{"col1\tcol2", "test\t", "\t"},
			str:    "\t",
			expect: true,
		},
		{
			name:   "quotes and escapes",
			array:  []string{"\"quoted\"", "'single'", "\\escaped"},
			str:    "\"quoted\"",
			expect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ArrContainsStr(tt.array, tt.str)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestArrContainsStr_UnicodeStrings(t *testing.T) {
	tests := []struct {
		name   string
		array  []string
		str    string
		expect bool
	}{
		{
			name:   "unicode - found",
			array:  []string{"hello", "世界", "🌍", "test"},
			str:    "世界",
			expect: true,
		},
		{
			name:   "unicode - not found",
			array:  []string{"hello", "world", "test"},
			str:    "世界",
			expect: false,
		},
		{
			name:   "emoji - found",
			array:  []string{"😀", "😂", "🎉", "🌟"},
			str:    "🎉",
			expect: true,
		},
		{
			name:   "emoji - not found",
			array:  []string{"😀", "😂", "🎉", "🌟"},
			str:    "❤️",
			expect: false,
		},
		{
			name:   "mixed unicode",
			array:  []string{"English", "中文", "日本語", "العربية", "🌍"},
			str:    "日本語",
			expect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ArrContainsStr(tt.array, tt.str)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestArrContainsStr_LongStrings(t *testing.T) {
	longString := "This is a very long string that contains many words and should test the function with longer text content that might be used in real applications."

	tests := []struct {
		name   string
		array  []string
		str    string
		expect bool
	}{
		{
			name:   "long string - found",
			array:  []string{"short", longString, "another"},
			str:    longString,
			expect: true,
		},
		{
			name:   "long string - not found",
			array:  []string{"short", "medium", "another"},
			str:    longString,
			expect: false,
		},
		{
			name:   "long string - partial match not found",
			array:  []string{"short", longString, "another"},
			str:    "This is a very long string",
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ArrContainsStr(tt.array, tt.str)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func BenchmarkArrContainsStr_SmallArray(b *testing.B) {
	array := []string{"apple", "banana", "cherry", "date", "elderberry"}
	str := "cherry"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ArrContainsStr(array, str)
	}
}

func BenchmarkArrContainsStr_MediumArray(b *testing.B) {
	array := make([]string, 100)
	for i := 0; i < 100; i++ {
		array[i] = "item" + string(rune(i))
	}
	str := "item50"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ArrContainsStr(array, str)
	}
}

func BenchmarkArrContainsStr_LargeArray(b *testing.B) {
	array := make([]string, 10000)
	for i := 0; i < 10000; i++ {
		array[i] = "item" + string(rune(i))
	}
	str := "item5000"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ArrContainsStr(array, str)
	}
}

func BenchmarkArrContainsStr_NotFound(b *testing.B) {
	array := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		array[i] = "item" + string(rune(i))
	}
	str := "notfound"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ArrContainsStr(array, str)
	}
}

func BenchmarkArrContainsStr_DuplicateStrings(b *testing.B) {
	array := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		array[i] = "duplicate"
	}
	str := "duplicate"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ArrContainsStr(array, str)
	}
}
