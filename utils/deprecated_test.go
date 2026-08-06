package utils

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	for i := range 10000 {
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
	for b.Loop() {
		ArrContainsStr(array, str)
	}
}

func BenchmarkArrContainsStr_MediumArray(b *testing.B) {
	array := make([]string, 100)
	for i := range 100 {
		array[i] = "item" + string(rune(i))
	}
	str := "item50"

	b.ResetTimer()
	for b.Loop() {
		ArrContainsStr(array, str)
	}
}

func BenchmarkArrContainsStr_LargeArray(b *testing.B) {
	array := make([]string, 10000)
	for i := range 10000 {
		array[i] = "item" + string(rune(i))
	}
	str := "item5000"

	b.ResetTimer()
	for b.Loop() {
		ArrContainsStr(array, str)
	}
}

func BenchmarkArrContainsStr_NotFound(b *testing.B) {
	array := make([]string, 1000)
	for i := range 1000 {
		array[i] = "item" + string(rune(i))
	}
	str := "notfound"

	b.ResetTimer()
	for b.Loop() {
		ArrContainsStr(array, str)
	}
}

func BenchmarkArrContainsStr_DuplicateStrings(b *testing.B) {
	array := make([]string, 1000)
	for i := range 1000 {
		array[i] = "duplicate"
	}
	str := "duplicate"

	b.ResetTimer()
	for b.Loop() {
		ArrContainsStr(array, str)
	}
}

func TestConvert_Success(t *testing.T) {
	tests := []struct {
		name          string
		value         string
		convertor     func(string) (int, error)
		expected      int
		expectedError bool
	}{
		{
			name:          "valid integer",
			value:         "123",
			convertor:     strconv.Atoi,
			expected:      123,
			expectedError: false,
		},
		{
			name:          "valid negative integer",
			value:         "-456",
			convertor:     strconv.Atoi,
			expected:      -456,
			expectedError: false,
		},
		{
			name:          "zero",
			value:         "0",
			convertor:     strconv.Atoi,
			expected:      0,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Convert(tt.value, tt.convertor)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expected, *result)
			}
		})
	}
}

func TestConvert_Error_NoDefault(t *testing.T) {
	result, err := Convert("invalid", strconv.Atoi)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestConvert_Error_WithDefault(t *testing.T) {
	defaultValue := 999
	result, err := Convert("invalid", strconv.Atoi, defaultValue)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, defaultValue, *result)
}

func TestConvert_Error_WithMultipleDefaults(t *testing.T) {
	// Should use the first default value
	defaultValue1 := 100
	defaultValue2 := 200
	result, err := Convert("invalid", strconv.Atoi, defaultValue1, defaultValue2)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, defaultValue1, *result)
}

func TestConvert_DifferentTypes(t *testing.T) {
	t.Run("string to float64", func(t *testing.T) {
		result, err := Convert("123.45", parseFloat64)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 123.45, *result)
	})

	t.Run("string to float64 with error and default", func(t *testing.T) {
		defaultValue := 99.99
		result, err := Convert("invalid", parseFloat64, defaultValue)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, defaultValue, *result)
	})

	t.Run("string to bool", func(t *testing.T) {
		result, err := Convert("true", strconv.ParseBool)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, true, *result)
	})

	t.Run("string to bool with error and default", func(t *testing.T) {
		defaultValue := false
		result, err := Convert("invalid", strconv.ParseBool, defaultValue)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, defaultValue, *result)
	})
}

func TestConvert_CustomConvertor(t *testing.T) {
	// Custom convertor that doubles the integer value
	doubleConvertor := func(s string) (int, error) {
		val, err := strconv.Atoi(s)
		if err != nil {
			return 0, err
		}
		return val * 2, nil
	}

	t.Run("custom convertor success", func(t *testing.T) {
		result, err := Convert("10", doubleConvertor)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 20, *result)
	})

	t.Run("custom convertor error with default", func(t *testing.T) {
		defaultValue := 50
		result, err := Convert("invalid", doubleConvertor, defaultValue)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, defaultValue, *result)
	})
}

func TestConvert_ErrorConvertor(t *testing.T) {
	// Convertor that always returns an error
	errorConvertor := func(s string) (string, error) {
		return "", errors.New("always fails")
	}

	t.Run("error convertor without default", func(t *testing.T) {
		result, err := Convert("test", errorConvertor)

		assert.Error(t, err)
		assert.Equal(t, "always fails", err.Error())
		assert.Nil(t, result)
	})

	t.Run("error convertor with default", func(t *testing.T) {
		defaultValue := "default"
		result, err := Convert("test", errorConvertor, defaultValue)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, defaultValue, *result)
	})
}

func TestConvert_StringToString(t *testing.T) {
	// Identity convertor for strings
	stringConvertor := func(s string) (string, error) {
		return s, nil
	}

	result, err := Convert("hello world", stringConvertor)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "hello world", *result)
}

func TestConvert_EmptyString(t *testing.T) {
	result, err := Convert("", strconv.Atoi)

	assert.Error(t, err) // strconv.Atoi("") returns an error
	assert.Nil(t, result)
}

func TestConvert_EmptyStringWithDefault(t *testing.T) {
	defaultValue := 42
	result, err := Convert("", strconv.Atoi, defaultValue)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, defaultValue, *result)
}

func TestConvert_ComplexType(t *testing.T) {
	// Test with a custom struct
	type Person struct {
		Name string
		Age  int
	}

	personConvertor := func(s string) (Person, error) {
		if s == "valid" {
			return Person{Name: "John", Age: 30}, nil
		}
		return Person{}, errors.New("invalid person")
	}

	t.Run("valid person", func(t *testing.T) {
		result, err := Convert("valid", personConvertor)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "John", result.Name)
		assert.Equal(t, 30, result.Age)
	})

	t.Run("invalid person with default", func(t *testing.T) {
		defaultPerson := Person{Name: "Default", Age: 0}
		result, err := Convert("invalid", personConvertor, defaultPerson)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "Default", result.Name)
		assert.Equal(t, 0, result.Age)
	})
}

func TestConvert_NilDefault(t *testing.T) {
	// Test with nil as default (though this would be unusual)
	result, err := Convert("invalid", strconv.Atoi, *new(int))

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 0, *result) // zero value of int
}

func BenchmarkConvert_Success(b *testing.B) {
	for b.Loop() {
		result, err := Convert("123", strconv.Atoi)
		if err != nil {
			b.Fatal(err)
		}
		if *result != 123 {
			b.Fatal("unexpected result")
		}
	}
}

func BenchmarkConvert_ErrorWithDefault(b *testing.B) {
	for b.Loop() {
		result, err := Convert("invalid", strconv.Atoi, 999)
		if err != nil {
			b.Fatal(err)
		}
		if *result != 999 {
			b.Fatal("unexpected result")
		}
	}
}

// Helper function to parse float64 (for testing purposes)
func parseFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// Helper function to test custom parsing
func parseHexInt(s string) (int, error) {
	if len(s) < 2 || s[:2] != "0x" {
		return 0, errors.New("not a hex number")
	}
	val, err := strconv.ParseInt(s[2:], 16, 32)
	return int(val), err
}

func TestConvert_HexExample(t *testing.T) {
	result, err := Convert("0xFF", parseHexInt)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 255, *result)
}

func TestConvert_HexExample_WithDefault(t *testing.T) {
	defaultValue := 0
	result, err := Convert("not_hex", parseHexInt, defaultValue)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, defaultValue, *result)
}

func TestCopyIOZeroAlloc(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "short string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "long string",
			input:    strings.Repeat("abcdefghijklmnopqrstuvwxyz", 1000),
			expected: strings.Repeat("abcdefghijklmnopqrstuvwxyz", 1000),
		},
		{
			name:     "string with special characters",
			input:    "Hello, 世界! 🌍 Special chars: !@#$%^&*()_+-=[]{}|;:,.<>?",
			expected: "Hello, 世界! 🌍 Special chars: !@#$%^&*()_+-=[]{}|;:,.<>?",
		},
		{
			name:     "multiline string",
			input:    "Line 1\nLine 2\nLine 3\n",
			expected: "Line 1\nLine 2\nLine 3\n",
		},
		{
			name:     "binary data",
			input:    string([]byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD}),
			expected: string([]byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			var writer bytes.Buffer

			n, err := CopyIOZeroAlloc(&writer, reader)

			require.NoError(t, err)
			assert.Equal(t, int64(len(tt.input)), n)
			assert.Equal(t, tt.expected, writer.String())
		})
	}
}

func TestCopyIOZeroAlloc_LargeData(t *testing.T) {
	// Test with data larger than buffer size (4096 bytes)
	largeData := strings.Repeat("This is a test string that will be repeated many times. ", 1000) // ~57KB
	reader := strings.NewReader(largeData)
	var writer bytes.Buffer

	n, err := CopyIOZeroAlloc(&writer, reader)

	require.NoError(t, err)
	assert.Equal(t, int64(len(largeData)), n)
	assert.Equal(t, largeData, writer.String())
}

func TestCopyIOZeroAlloc_ErrorReader(t *testing.T) {
	// Test with reader that returns an error
	errorReader := &errorReader{err: io.ErrUnexpectedEOF}
	var writer bytes.Buffer

	n, err := CopyIOZeroAlloc(&writer, errorReader)

	assert.Equal(t, int64(0), n)
	assert.Equal(t, io.ErrUnexpectedEOF, err)
}

func TestCopyIOZeroAlloc_ErrorWriter(t *testing.T) {
	// Test with writer that returns an error
	reader := strings.NewReader("test data")
	errorWriter := &errorWriter{err: io.ErrShortWrite}

	n, err := CopyIOZeroAlloc(errorWriter, reader)

	assert.Equal(t, int64(0), n)
	assert.Equal(t, io.ErrShortWrite, err)
}

func TestCopyIOZeroAlloc_BufferPoolReuse(t *testing.T) {
	// Test that buffer pool is properly reused
	reader1 := strings.NewReader("first test")
	var writer1 bytes.Buffer

	n1, err1 := CopyIOZeroAlloc(&writer1, reader1)
	require.NoError(t, err1)
	assert.Equal(t, int64(10), n1)
	assert.Equal(t, "first test", writer1.String())

	reader2 := strings.NewReader("second test")
	var writer2 bytes.Buffer

	n2, err2 := CopyIOZeroAlloc(&writer2, reader2)
	require.NoError(t, err2)
	assert.Equal(t, int64(11), n2)
	assert.Equal(t, "second test", writer2.String())
}

func TestCopyIOZeroAlloc_ConcurrentAccess(t *testing.T) {
	// Test concurrent access to buffer pool
	const numGoroutines = 10
	const iterations = 100

	done := make(chan bool, numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			for range iterations {
				data := strings.Repeat("test", id+1)
				reader := strings.NewReader(data)
				var writer bytes.Buffer

				n, err := CopyIOZeroAlloc(&writer, reader)

				assert.NoError(t, err)
				assert.Equal(t, int64(len(data)), n)
				assert.Equal(t, data, writer.String())
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for range numGoroutines {
		<-done
	}
}

func TestCopyIOZeroAlloc_EmptyReader(t *testing.T) {
	reader := strings.NewReader("")
	var writer bytes.Buffer

	n, err := CopyIOZeroAlloc(&writer, reader)

	require.NoError(t, err)
	assert.Equal(t, int64(0), n)
	assert.Equal(t, "", writer.String())
}

func BenchmarkCopyIOZeroAlloc(b *testing.B) {
	data := strings.Repeat("benchmark test data ", 1000) // ~20KB

	b.ResetTimer()
	for b.Loop() {
		reader := strings.NewReader(data)
		var writer bytes.Buffer

		_, err := CopyIOZeroAlloc(&writer, reader)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCopyIOZeroAlloc_vs_ioCopy(b *testing.B) {
	data := strings.Repeat("benchmark test data ", 1000) // ~20KB

	b.Run("CopyIOZeroAlloc", func(b *testing.B) {
		for b.Loop() {
			reader := strings.NewReader(data)
			var writer bytes.Buffer

			_, err := CopyIOZeroAlloc(&writer, reader)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("io.Copy", func(b *testing.B) {
		for b.Loop() {
			reader := strings.NewReader(data)
			var writer bytes.Buffer

			_, err := io.Copy(&writer, reader)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// Helper types for testing error conditions
type errorReader struct {
	err error
}

func (er *errorReader) Read(p []byte) (n int, err error) {
	return 0, er.err
}

type errorWriter struct {
	err error
}

func (ew *errorWriter) Write(p []byte) (n int, err error) {
	return 0, ew.err
}

func TestFilePathToIOReader_Success(t *testing.T) {
	// Create a temporary file
	tmpDir, err := os.MkdirTemp("", "file-io-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testContent := "Hello, World!\nThis is a test file.\n"
	testFile := filepath.Join(tmpDir, "test.txt")

	err = os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err)

	// Test the function
	reader, err := FilePathToIOReader(testFile)
	require.NoError(t, err)
	require.NotNil(t, reader)

	// Read the content
	content, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, testContent, string(content))

	// Close the file (reader should be a *os.File)
	if file, ok := reader.(*os.File); ok {
		_ = file.Close()
	}
}

func TestFilePathToIOReader_EmptyFile(t *testing.T) {
	// Create a temporary empty file
	tmpDir, err := os.MkdirTemp("", "file-io-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "empty.txt")
	err = os.WriteFile(testFile, []byte(""), 0644)
	require.NoError(t, err)

	// Test the function
	reader, err := FilePathToIOReader(testFile)
	require.NoError(t, err)
	require.NotNil(t, reader)

	// Read the content
	content, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, "", string(content))

	// Close the file
	if file, ok := reader.(*os.File); ok {
		_ = file.Close()
	}
}

func TestFilePathToIOReader_LargeFile(t *testing.T) {
	// Create a temporary large file
	tmpDir, err := os.MkdirTemp("", "file-io-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a large content (1MB)
	largeContent := strings.Repeat("This is a line of text for testing large files.\n", 20000)
	testFile := filepath.Join(tmpDir, "large.txt")

	err = os.WriteFile(testFile, []byte(largeContent), 0644)
	require.NoError(t, err)

	// Test the function
	reader, err := FilePathToIOReader(testFile)
	require.NoError(t, err)
	require.NotNil(t, reader)

	// Read the content
	content, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, largeContent, string(content))

	// Close the file
	if file, ok := reader.(*os.File); ok {
		_ = file.Close()
	}
}

func TestFilePathToIOReader_BinaryFile(t *testing.T) {
	// Create a temporary binary file
	tmpDir, err := os.MkdirTemp("", "file-io-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create binary content
	binaryContent := make([]byte, 256)
	for i := range binaryContent {
		binaryContent[i] = byte(i)
	}

	testFile := filepath.Join(tmpDir, "binary.bin")
	err = os.WriteFile(testFile, binaryContent, 0644)
	require.NoError(t, err)

	// Test the function
	reader, err := FilePathToIOReader(testFile)
	require.NoError(t, err)
	require.NotNil(t, reader)

	// Read the content
	content, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, binaryContent, content)

	// Close the file
	if file, ok := reader.(*os.File); ok {
		_ = file.Close()
	}
}

func TestFilePathToIOReader_FileNotFound(t *testing.T) {
	// Test with non-existent file
	nonExistentFile := "/path/to/nonexistent/file.txt"

	reader, err := FilePathToIOReader(nonExistentFile)
	assert.Error(t, err)
	assert.Nil(t, reader)
	assert.True(t, os.IsNotExist(err))
}

func TestFilePathToIOReader_Directory(t *testing.T) {
	// Test with a directory instead of a file
	tmpDir, err := os.MkdirTemp("", "file-io-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	reader, err := FilePathToIOReader(tmpDir)
	// Note: On some systems, opening a directory might succeed but fail when reading
	// We'll just verify that if it succeeds, it's still a valid reader
	if err != nil {
		assert.Nil(t, reader)
	} else {
		assert.NotNil(t, reader)
		// Close it if it's a file
		if file, ok := reader.(*os.File); ok {
			_ = file.Close()
		}
	}
}

func TestFilePathToIOReader_PermissionDenied(t *testing.T) {
	// Skip this test on Windows as permission handling is different
	if os.Getenv("OS") == "Windows_NT" {
		t.Skip("Skipping permission test on Windows")
	}

	// Create a temporary file with no read permissions
	tmpDir, err := os.MkdirTemp("", "file-io-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "no-read.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Remove read permissions
	err = os.Chmod(testFile, 0000)
	require.NoError(t, err)

	// Restore permissions after test for cleanup
	defer func() { _ = os.Chmod(testFile, 0644) }()

	reader, err := FilePathToIOReader(testFile)
	assert.Error(t, err)
	assert.Nil(t, reader)
}

func TestFilePathToIOReader_EmptyPath(t *testing.T) {
	// Test with empty path
	reader, err := FilePathToIOReader("")
	assert.Error(t, err)
	assert.Nil(t, reader)
}

func TestFilePathToIOReader_ReaderIsFile(t *testing.T) {
	// Test that the returned reader is actually a *os.File
	tmpDir, err := os.MkdirTemp("", "file-io-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	reader, err := FilePathToIOReader(testFile)
	require.NoError(t, err)
	require.NotNil(t, reader)

	// Check that it's a *os.File
	file, ok := reader.(*os.File)
	assert.True(t, ok, "Reader should be a *os.File")
	assert.NotNil(t, file)

	// Test file operations
	stat, err := file.Stat()
	require.NoError(t, err)
	assert.Equal(t, "test.txt", stat.Name())
	assert.Equal(t, int64(4), stat.Size())

	_ = file.Close()
}

func TestFilePathToIOReader_MultipleReads(t *testing.T) {
	// Test that the reader can be used for multiple reads
	tmpDir, err := os.MkdirTemp("", "file-io-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testContent := "Line 1\nLine 2\nLine 3\n"
	testFile := filepath.Join(tmpDir, "multiread.txt")
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err)

	reader, err := FilePathToIOReader(testFile)
	require.NoError(t, err)
	require.NotNil(t, reader)

	// Read first part
	buffer1 := make([]byte, 7)
	n1, err := reader.Read(buffer1)
	require.NoError(t, err)
	assert.Equal(t, 7, n1)
	assert.Equal(t, "Line 1\n", string(buffer1))

	// Read second part
	buffer2 := make([]byte, 7)
	n2, err := reader.Read(buffer2)
	require.NoError(t, err)
	assert.Equal(t, 7, n2)
	assert.Equal(t, "Line 2\n", string(buffer2))

	// Read remaining part
	remaining, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, "Line 3\n", string(remaining))

	// Close the file
	if file, ok := reader.(*os.File); ok {
		_ = file.Close()
	}
}

func TestFilePathToIOReader_ReadAtEOF(t *testing.T) {
	// Test reading at EOF
	tmpDir, err := os.MkdirTemp("", "file-io-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "eof.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	reader, err := FilePathToIOReader(testFile)
	require.NoError(t, err)
	require.NotNil(t, reader)

	// Read all content
	content, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, "test", string(content))

	// Try to read again - should return EOF
	buffer := make([]byte, 10)
	n, err := reader.Read(buffer)
	assert.Equal(t, 0, n)
	assert.Equal(t, io.EOF, err)

	// Close the file
	if file, ok := reader.(*os.File); ok {
		_ = file.Close()
	}
}

func TestFilePathToIOReader_AbsolutePath(t *testing.T) {
	// Test with absolute path
	tmpDir, err := os.MkdirTemp("", "file-io-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testFile := filepath.Join(tmpDir, "absolute.txt")
	err = os.WriteFile(testFile, []byte("absolute path test"), 0644)
	require.NoError(t, err)

	// Get absolute path
	absPath, err := filepath.Abs(testFile)
	require.NoError(t, err)

	reader, err := FilePathToIOReader(absPath)
	require.NoError(t, err)
	require.NotNil(t, reader)

	content, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, "absolute path test", string(content))

	// Close the file
	if file, ok := reader.(*os.File); ok {
		_ = file.Close()
	}
}

func TestFilePathToIOReader_RelativePath(t *testing.T) {
	// Test with relative path
	tmpDir, err := os.MkdirTemp("", "file-io-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	testFile := "relative.txt"
	err = os.WriteFile(testFile, []byte("relative path test"), 0644)
	require.NoError(t, err)

	reader, err := FilePathToIOReader(testFile)
	require.NoError(t, err)
	require.NotNil(t, reader)

	content, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, "relative path test", string(content))

	// Close the file
	if file, ok := reader.(*os.File); ok {
		_ = file.Close()
	}
}

func BenchmarkFilePathToIOReader(b *testing.B) {
	// Create a temporary file
	tmpDir, err := os.MkdirTemp("", "file-io-benchmark-*")
	require.NoError(b, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testContent := strings.Repeat("benchmark test data\n", 1000)
	testFile := filepath.Join(tmpDir, "benchmark.txt")
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(b, err)

	b.ResetTimer()
	for b.Loop() {
		reader, err := FilePathToIOReader(testFile)
		if err != nil {
			b.Fatal(err)
		}

		// Read some content to make the benchmark more realistic
		buffer := make([]byte, 1024)
		_, err = reader.Read(buffer)
		if err != nil && err != io.EOF {
			b.Fatal(err)
		}

		// Close the file
		if file, ok := reader.(*os.File); ok {
			_ = file.Close()
		}
	}
}
