package utils

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < iterations; j++ {
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
	for i := 0; i < numGoroutines; i++ {
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
	for i := 0; i < b.N; i++ {
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
		for i := 0; i < b.N; i++ {
			reader := strings.NewReader(data)
			var writer bytes.Buffer

			_, err := CopyIOZeroAlloc(&writer, reader)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("io.Copy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
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
