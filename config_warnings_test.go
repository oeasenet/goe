package goe_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2"
)

// A value that does not parse — a duration without a unit — used to become
// zero silently. It is read before GOE's logger exists, and must still be
// reported once.
func TestConfig_UnparsableValuesAreReportedAtStartup(t *testing.T) {
	if os.Getenv("GOE_CONFIG_HELPER") == "1" {
		goe.New(goe.Options{})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestConfig_UnparsableValuesAreReportedAtStartup$")
	cmd.Env = append(os.Environ(), "GOE_CONFIG_HELPER=1", "SHUTDOWN_TIMEOUT=30", "LOG_FORMAT=json")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	var warnings []string
	for line := range strings.SplitSeq(string(out), "\n") {
		if strings.Contains(line, "Config value does not parse") {
			warnings = append(warnings, line)
		}
	}
	require.Len(t, warnings, 1, string(out))
	assert.Contains(t, warnings[0], `"key":"SHUTDOWN_TIMEOUT"`)
	assert.Contains(t, warnings[0], `"type":"duration"`)
}
