package goe_test

import (
	"errors"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2"
)

// MongoDB connects while goe.New builds the app, like cache, job and lock, so
// DB() is ready for providers and invokers and an unreachable database stops
// the process there.
func TestMongoDB_UnreachableFailsInsideNew(t *testing.T) {
	if os.Getenv("GOE_MONGO_HELPER") == "1" {
		goe.New(goe.Options{
			WithMongoDB: true,
			ConfigOverrides: map[string]any{
				"MONGO_URI":          "mongodb://127.0.0.1:1/?serverSelectionTimeoutMS=200&directConnection=true",
				"MONGO_DB_NAME":      "goe_connect_test",
				"MONGO_PING_TIMEOUT": "300ms",
			},
		})
		os.Exit(0) // reached only when New did not connect
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestMongoDB_UnreachableFailsInsideNew$")
	cmd.Env = append(os.Environ(), "GOE_MONGO_HELPER=1")
	start := time.Now()
	out, err := cmd.CombinedOutput()

	var exitErr *exec.ExitError
	require.True(t, errors.As(err, &exitErr), "goe.New must stop the process, got %v\n%s", err, out)
	assert.Contains(t, string(out), "Failed to connect to MongoDB")
	assert.Less(t, time.Since(start), 10*time.Second)
}
