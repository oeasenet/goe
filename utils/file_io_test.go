package utils

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	for i := 0; i < b.N; i++ {
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
