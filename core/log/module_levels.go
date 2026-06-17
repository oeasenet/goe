package log

import (
	"strings"

	"go.uber.org/zap/zapcore"
)

// moduleFieldKey is the structured-log field key identifying which module a log
// entry belongs to. Modules tag their logger once via With(moduleFieldKey, name).
const moduleFieldKey = "module"

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

	for _, pair := range strings.Split(raw, ",") {
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
