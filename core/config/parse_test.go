package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.oease.dev/goe/v2/contract"
)

// warnRecorder records Warn calls; other methods are never called here.
type warnRecorder struct {
	contract.Logger
	mu    sync.Mutex
	warns []map[string]any
}

func (r *warnRecorder) Warn(msg string, args ...any) {
	fields := map[string]any{"msg": msg}
	for i := 0; i+1 < len(args); i += 2 {
		fields[args[i].(string)] = args[i+1]
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.warns = append(r.warns, fields)
}

func configWith(values map[string]any) *config {
	return &config{data: values, cache: map[string]any{}}
}

func TestGetDuration_WarnsOnceForAnUnparsableValue(t *testing.T) {
	c := configWith(map[string]any{"SHUTDOWN_TIMEOUT": "30"})
	rec := &warnRecorder{}
	AttachLogger(c, rec)

	assert.Zero(t, c.GetDuration("SHUTDOWN_TIMEOUT"), "the default is returned as before")
	assert.Zero(t, c.GetDuration("SHUTDOWN_TIMEOUT"))

	require.Len(t, rec.warns, 1, "one warning per key")
	assert.Equal(t, "Config value does not parse; using the default", rec.warns[0]["msg"])
	assert.Equal(t, "SHUTDOWN_TIMEOUT", rec.warns[0]["key"])
	assert.Equal(t, "30", rec.warns[0]["value"])
	assert.Equal(t, "duration", rec.warns[0]["type"])
}

func TestGetters_WarnForEveryNumericAndBooleanType(t *testing.T) {
	c := configWith(map[string]any{"I": "ten", "I64": "x", "F": "1,5", "B": "yes"})
	rec := &warnRecorder{}
	AttachLogger(c, rec)
	c.GetInt("I")
	c.GetInt64("I64")
	c.GetFloat64("F")
	c.GetBool("B")

	var types []string
	for _, w := range rec.warns {
		types = append(types, w["type"].(string))
	}
	assert.ElementsMatch(t, []string{"int", "int64", "float64", "bool"}, types)
}

func TestGetters_WarningRecordedBeforeAttachIsFlushed(t *testing.T) {
	c := configWith(map[string]any{"LOCK_TTL": "5 minutes"})
	assert.Zero(t, c.GetDuration("LOCK_TTL"), "read before any logger exists")

	rec := &warnRecorder{}
	AttachLogger(c, rec)
	require.Len(t, rec.warns, 1)
	assert.Equal(t, "LOCK_TTL", rec.warns[0]["key"])
}

func TestGetters_NoWarningForUnsetOrValidValues(t *testing.T) {
	c := configWith(map[string]any{"PORT": "8080", "ON": "true", "WAIT": "2s"})
	rec := &warnRecorder{}
	AttachLogger(c, rec)
	assert.Equal(t, 8080, c.GetInt("PORT"))
	assert.True(t, c.GetBool("ON"))
	assert.Zero(t, c.GetInt("MISSING"))
	assert.Empty(t, rec.warns)
}

func TestAttachLogger_ReachesTheWrappedConfig(t *testing.T) {
	base := configWith(map[string]any{"N": "nope"})
	wrapped := NewConfigWrapper(base, map[string]any{"OTHER": 1})
	rec := &warnRecorder{}
	AttachLogger(wrapped, rec)
	wrapped.GetInt("N")
	assert.Len(t, rec.warns, 1)
}

func loadEnv(t *testing.T, content string) map[string]any {
	t.Helper()
	file := filepath.Join(t.TempDir(), ".env")
	require.NoError(t, os.WriteFile(file, []byte(content), 0o600))
	c := configWith(map[string]any{})
	c.loadEnvFile(file)
	return c.data
}

func TestLoadEnvFile_ExistingFormsUnchanged(t *testing.T) {
	data := loadEnv(t, "A=plain\n"+
		"B = spaced \n"+
		"C=\"quoted\"\n"+
		"D='single'\n"+
		"E=\n"+
		"F=a=b\n"+
		"# comment\n"+
		"G=with # hash\n"+
		"I=trailing\"\n"+
		"J=  \"padded\"  \n"+
		"K=C:\\new\\table\n"+
		"H=\"unclosed\n") // an unclosed quote at the end of the file
	assert.Equal(t, map[string]any{
		"A": "plain", "B": "spaced", "C": "quoted", "D": "single", "E": "",
		"F": "a=b", "G": "with # hash", "H": "unclosed", "I": "trailing",
		"J": "padded", "K": `C:\new\table`,
	}, data)
}

func TestLoadEnvFile_MultiLineDoubleQuoted(t *testing.T) {
	data := loadEnv(t, "KEY=\"-----BEGIN PRIVATE KEY-----\n"+
		"MIIBVQIBADANBgkqhkiG\n"+
		"-----END PRIVATE KEY-----\"\n"+
		"NEXT=after\n")
	assert.Equal(t, "-----BEGIN PRIVATE KEY-----\nMIIBVQIBADANBgkqhkiG\n-----END PRIVATE KEY-----", data["KEY"])
	assert.Equal(t, "after", data["NEXT"], "parsing continues after the value")
}

func TestLoadEnvFile_DoubleQuotedEscapes(t *testing.T) {
	data := loadEnv(t, `PEM="line1\nline2"`+"\n"+`Q="say \"hi\""`+"\n"+`S="back\\slash\tTab"`+"\n")
	assert.Equal(t, "line1\nline2", data["PEM"])
	assert.Equal(t, `say "hi"`, data["Q"])
	assert.Equal(t, "back\\slash\tTab", data["S"])
}

func TestLoadEnvFile_SingleQuotedIsLiteral(t *testing.T) {
	data := loadEnv(t, `RAW='a\nb'`+"\n"+"MULTI='first\nsecond'\n")
	assert.Equal(t, `a\nb`, data["RAW"])
	assert.Equal(t, "first\nsecond", data["MULTI"])
}

const unbalancedQuotes = "Env file value has unbalanced quotes; read without its outer quotes"

func loadEnvConfig(t *testing.T, content string) (*config, *warnRecorder) {
	t.Helper()
	file := filepath.Join(t.TempDir(), ".env")
	require.NoError(t, os.WriteFile(file, []byte(content), 0o600))
	c := configWith(map[string]any{})
	c.loadEnvFile(file)
	rec := &warnRecorder{}
	AttachLogger(c, rec)
	return c, rec
}

func warnedKeys(rec *warnRecorder, msg string) []string {
	var keys []string
	for _, w := range rec.warns {
		if w["msg"] == msg {
			keys = append(keys, w["key"].(string))
		}
	}
	return keys
}

// A quote that is never closed must not swallow the keys after it.
func TestLoadEnvFile_UnclosedQuoteDoesNotSwallowKeys(t *testing.T) {
	c, rec := loadEnvConfig(t, "A=\"oops\nB=1\nC=\"fine\"\nG='Hello\nX=2\nNAME=o'brien\n")
	assert.Equal(t, map[string]any{
		"A": "oops", "B": "1", "C": "fine", "G": "Hello", "X": "2", "NAME": "o'brien",
	}, c.data)
	assert.ElementsMatch(t, []string{"A", "G"}, warnedKeys(rec, unbalancedQuotes))
}

// Text after the closing quote means the line was never a quoted value: it
// reads as it always did, so an existing secret does not change.
func TestLoadEnvFile_TextAfterTheClosingQuoteReadsAsBefore(t *testing.T) {
	c, rec := loadEnvConfig(t, `PASSWORD="p@ss"word"`+"\n"+`D='don't'`+"\n"+`J="{"k":"v"}"`+"\n"+
		`H="abc"#def`+"\n") // a comment needs a space before it
	assert.Equal(t, `p@ss"word`, c.data["PASSWORD"])
	assert.Equal(t, `don't`, c.data["D"])
	assert.Equal(t, `{"k":"v"}`, c.data["J"])
	assert.Equal(t, `abc"#def`, c.data["H"])
	assert.ElementsMatch(t, []string{"PASSWORD", "D", "J", "H"}, warnedKeys(rec, unbalancedQuotes))
}

func TestLoadEnvFile_CommentAfterTheClosingQuote(t *testing.T) {
	c, rec := loadEnvConfig(t, `KEY="value" # where it comes from`+"\n")
	assert.Equal(t, "value", c.data["KEY"])
	assert.Empty(t, rec.warns)
}
