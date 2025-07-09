package utils

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	for i := 0; i < b.N; i++ {
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
	for i := 0; i < b.N; i++ {
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
