package log

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loggerWithOutput(output ...string) *defaultLoggerConfig {
	return &defaultLoggerConfig{level: "info", format: "json", output: output}
}

func TestNew_StdoutIsConsole(t *testing.T) {
	out := captureStdout(t, func() {
		New(loggerWithOutput("stdout")).Info("hello")
	})
	assert.Contains(t, out, "hello", "LOG_OUTPUT=stdout used to drop every line")
}

func TestNew_UnknownLogOutputFallsBackToConsole(t *testing.T) {
	out := captureStdout(t, func() {
		New(loggerWithOutput("stdot")).Info("hello")
	})
	assert.Contains(t, out, "hello", "logs are never dropped for a mistyped output")
	assert.Contains(t, out, "LOG_OUTPUT value is not recognised; logging to console")
	assert.Contains(t, out, "stdot")
}

func TestNew_ConsoleIsWrittenOnce(t *testing.T) {
	out := captureStdout(t, func() {
		New(loggerWithOutput("console", "stdout", "nope")).Info("hello")
	})
	assert.Equal(t, 1, strings.Count(out, `"msg":"hello"`))
}

func TestNew_StderrOutput(t *testing.T) {
	old := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w
	New(loggerWithOutput("stderr")).Info("to-stderr")
	os.Stderr = old
	require.NoError(t, w.Close())
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	assert.Contains(t, buf.String(), "to-stderr")
}
