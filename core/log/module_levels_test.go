package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestParseLevel(t *testing.T) {
	cases := map[string]struct {
		want zapcore.Level
		ok   bool
	}{
		"debug":   {zapcore.DebugLevel, true},
		"INFO":    {zapcore.InfoLevel, true},
		" warn ":  {zapcore.WarnLevel, true},
		"warning": {zapcore.WarnLevel, true},
		"error":   {zapcore.ErrorLevel, true},
		"nope":    {zapcore.InfoLevel, false},
		"":        {zapcore.InfoLevel, false},
	}
	for in, exp := range cases {
		got, ok := parseLevel(in)
		if got != exp.want || ok != exp.ok {
			t.Errorf("parseLevel(%q) = (%v,%v), want (%v,%v)", in, got, ok, exp.want, exp.ok)
		}
	}
}

func TestParseModuleLevels(t *testing.T) {
	levels, invalid := parseModuleLevels(" job:debug, gorm:warn ,JOB2:ERROR ")
	if len(invalid) != 0 {
		t.Fatalf("unexpected invalid entries: %v", invalid)
	}
	if levels["job"] != zapcore.DebugLevel {
		t.Errorf("job = %v, want debug", levels["job"])
	}
	if levels["gorm"] != zapcore.WarnLevel {
		t.Errorf("gorm = %v, want warn", levels["gorm"])
	}
	if levels["job2"] != zapcore.ErrorLevel { // module name lower-cased
		t.Errorf("job2 = %v, want error", levels["job2"])
	}
}

func TestParseModuleLevels_SkipsMalformed(t *testing.T) {
	levels, invalid := parseModuleLevels("bad,job:nope,:warn,http:info,")
	if levels["http"] != zapcore.InfoLevel {
		t.Errorf("http = %v, want info", levels["http"])
	}
	if _, ok := levels["job"]; ok {
		t.Error("job should be skipped (invalid level)")
	}
	if len(invalid) != 3 { // "bad", "job:nope", ":warn"
		t.Errorf("invalid = %v, want 3 entries", invalid)
	}
}

func TestParseModuleLevels_Empty(t *testing.T) {
	levels, invalid := parseModuleLevels("")
	if len(levels) != 0 || len(invalid) != 0 {
		t.Errorf("empty input should yield empty results, got %v / %v", levels, invalid)
	}
}

func TestModuleLevels_LevelFor(t *testing.T) {
	m := &moduleLevels{
		global:    zapcore.InfoLevel,
		overrides: map[string]zapcore.Level{"job": zapcore.DebugLevel, "gorm": zapcore.WarnLevel},
	}
	if m.levelFor("") != zapcore.InfoLevel {
		t.Error("root module should use global")
	}
	if m.levelFor("unknown") != zapcore.InfoLevel {
		t.Error("unknown module should use global")
	}
	if m.levelFor("job") != zapcore.DebugLevel {
		t.Error("job override not applied")
	}
	if m.levelFor("gorm") != zapcore.WarnLevel {
		t.Error("gorm override not applied")
	}
}

func TestWithDefaultModuleLevels(t *testing.T) {
	t.Run("fx defaults to warn so wiring narration is suppressed", func(t *testing.T) {
		got := withDefaultModuleLevels(nil)
		assert.Equal(t, zapcore.WarnLevel, got[fxModuleName])
	})

	t.Run("explicit override wins over the default", func(t *testing.T) {
		got := withDefaultModuleLevels(map[string]zapcore.Level{
			fxModuleName: zapcore.DebugLevel,
		})
		assert.Equal(t, zapcore.DebugLevel, got[fxModuleName],
			"LOG_MODULE_LEVELS=fx:debug must re-enable Fx detail")
	})

	t.Run("unrelated overrides are preserved", func(t *testing.T) {
		got := withDefaultModuleLevels(map[string]zapcore.Level{
			"job": zapcore.DebugLevel,
		})
		assert.Equal(t, zapcore.DebugLevel, got["job"])
		assert.Equal(t, zapcore.WarnLevel, got[fxModuleName])
	})

	t.Run("fx errors still pass the default gate", func(t *testing.T) {
		levels := &moduleLevels{
			global:    zapcore.InfoLevel,
			overrides: withDefaultModuleLevels(nil),
		}
		assert.Equal(t, zapcore.WarnLevel, levels.levelFor(fxModuleName))
		assert.True(t, zapcore.ErrorLevel >= levels.levelFor(fxModuleName),
			"Fx failures must still be logged")
		assert.False(t, zapcore.DebugLevel >= levels.levelFor(fxModuleName),
			"Fx debug narration must be dropped")
	})
}
