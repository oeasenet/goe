package log

import (
	"strings"

	"go.uber.org/zap/zapcore"
)

// moduleFieldKey is the structured-log field key identifying which module a log
// entry belongs to. Modules tag their logger once via With(moduleFieldKey, name).
const moduleFieldKey = "module"

// fxModuleName tags log entries produced by the Uber Fx event logger.
const fxModuleName = "fx"

// defaultModuleLevels are per-module levels applied when the operator has not
// set one explicitly in LOG_MODULE_LEVELS.
//
// Fx emits an event for every constructor supplied, provided, decorated and run,
// plus each lifecycle hook. That is dozens of lines of dependency-graph detail
// which is only interesting when diagnosing wiring itself, and previously it was
// tied to the global level — so raising LOG_LEVEL to debug for your own code
// buried it. Fx now defaults to warn, keeping its failures (which is what
// actually matters when a provide or invoke breaks) and dropping the narration.
//
// Opt back in with LOG_MODULE_LEVELS=fx:debug.
var defaultModuleLevels = map[string]zapcore.Level{
	fxModuleName: zapcore.WarnLevel,
}

// withDefaultModuleLevels returns overrides with the built-in defaults filled in
// for any module the operator did not configure. Explicit settings always win.
func withDefaultModuleLevels(overrides map[string]zapcore.Level) map[string]zapcore.Level {
	if overrides == nil {
		overrides = make(map[string]zapcore.Level, len(defaultModuleLevels))
	}
	for module, level := range defaultModuleLevels {
		if _, set := overrides[module]; !set {
			overrides[module] = level
		}
	}
	return overrides
}

// moduleLevels holds the global (root) log level plus optional per-module
// overrides. It is built once at startup and never mutated afterwards, so it is
// safe for concurrent reads without locking.
type moduleLevels struct {
	global    zapcore.Level
	overrides map[string]zapcore.Level
}

// levelFor returns the minimum enabled level for the given module name. The
// empty/root module and any module without an override fall back to global.
func (m *moduleLevels) levelFor(module string) zapcore.Level {
	if module != "" && m.overrides != nil {
		if lvl, ok := m.overrides[module]; ok {
			return lvl
		}
	}
	return m.global
}

// parseLevel converts a level word into a zapcore.Level. The bool reports
// whether the word was recognized. Unknown words map to Info (and false).
func parseLevel(s string) (zapcore.Level, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return zapcore.DebugLevel, true
	case "info":
		return zapcore.InfoLevel, true
	case "warn", "warning":
		return zapcore.WarnLevel, true
	case "error":
		return zapcore.ErrorLevel, true
	default:
		return zapcore.InfoLevel, false
	}
}

// parseModuleLevels parses a comma-separated "module:level" spec such as
// "job:debug,gorm:warn" into a level map. Module names are lower-cased. Malformed
// entries are skipped and returned in the second slice so the caller can warn.
// It never panics on bad input.
func parseModuleLevels(raw string) (map[string]zapcore.Level, []string) {
	result := make(map[string]zapcore.Level)
	var invalid []string

	for pair := range strings.SplitSeq(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		name, lvlStr, ok := strings.Cut(pair, ":")
		name = strings.ToLower(strings.TrimSpace(name))
		lvlStr = strings.TrimSpace(lvlStr)
		if !ok || name == "" || lvlStr == "" {
			invalid = append(invalid, pair)
			continue
		}

		lvl, valid := parseLevel(lvlStr)
		if !valid {
			invalid = append(invalid, pair)
			continue
		}
		result[name] = lvl
	}

	return result, invalid
}
