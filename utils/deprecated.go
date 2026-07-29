// Helpers kept for backward compatibility. Each has a better replacement in the
// standard library, noted on the function. They still work and are still tested;
// they are simply no longer the recommended way to do these things.

package utils

import (
	"io"
	"os"
	"sync"
)

// ArrContainsStr reports whether str is present in array.
//
// Deprecated: use slices.Contains(array, str) from the standard library. This
// implementation builds a map of every element before doing a single lookup,
// which allocates for no benefit.
func ArrContainsStr(array []string, str string) bool {
	m := make(map[string]struct{}, len(array))
	for _, s := range array {
		m[s] = struct{}{}
	}
	_, exists := m[str]
	return exists
}

// Convert applies convertor to value, falling back to the first defaultValue if
// the conversion fails.
//
// Deprecated: call the conversion function directly. This wrapper returns a
// pointer where a value would do, and silently discards the conversion error
// whenever a default is supplied, which hides malformed input.
func Convert[T any](value string, convertor func(string) (T, error), defaultValue ...T) (*T, error) {
	result, err := convertor(value)
	if err != nil {
		if len(defaultValue) > 0 {
			return &defaultValue[0], nil
		}
		return nil, err
	}
	return &result, nil
}

var copyBufPool = sync.Pool{
	New: func() any {
		return make([]byte, 4096)
	},
}

// CopyIOZeroAlloc copies r into w using a pooled buffer.
//
// Deprecated: use io.Copy. The name overstates the benefit — this pools a buffer
// rather than avoiding allocation — and its 4KiB buffer is smaller than the
// 32KiB io.Copy uses, so it is typically slower on large payloads.
func CopyIOZeroAlloc(w io.Writer, r io.Reader) (int64, error) {
	vbuf := copyBufPool.Get()
	buf := vbuf.([]byte)
	n, err := io.CopyBuffer(w, r, buf)
	copyBufPool.Put(vbuf)
	return n, err
}

// FilePathToIOReader opens filePath and returns it as an io.Reader.
//
// Deprecated: use os.Open. Returning io.Reader discards the *os.File, so the
// caller cannot Close it and the descriptor leaks for the lifetime of the
// process. os.Open returns a value you can defer Close on.
func FilePathToIOReader(filePath string) (io.Reader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}
