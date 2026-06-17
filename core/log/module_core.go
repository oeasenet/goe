package log

import "go.uber.org/zap/zapcore"

// moduleLevelCore wraps a zapcore.Core and gates entries by a per-module level.
// The module name is captured from the "module" structured field as it flows
// through With(...), so a logger tagged with With("module","job") gets its own
// effective level from the shared moduleLevels table.
//
// This makes LOG_LEVEL a global default while LOG_MODULE_LEVELS raises or lowers
// individual modules independently. The wrapped (inner) core should be enabled at
// the most permissive level (Debug); this wrapper performs all real filtering and
// delegates Write/Sync to the inner core unchanged.
type moduleLevelCore struct {
	zapcore.Core
	module string
	levels *moduleLevels
}

// newModuleLevelCore wraps inner with per-module gating.
func newModuleLevelCore(inner zapcore.Core, levels *moduleLevels) *moduleLevelCore {
	return &moduleLevelCore{Core: inner, levels: levels}
}

// Enabled reports whether lvl passes the gate for this core's module.
func (c *moduleLevelCore) Enabled(lvl zapcore.Level) bool {
	return lvl >= c.levels.levelFor(c.module)
}

// With returns a child core, updating the tracked module name when the fields
// carry a "module" string field. Field application is delegated to the inner core
// so the field is still emitted normally.
func (c *moduleLevelCore) With(fields []zapcore.Field) zapcore.Core {
	module := c.module
	for _, f := range fields {
		if f.Key == moduleFieldKey && f.Type == zapcore.StringType {
			module = f.String
		}
	}
	return &moduleLevelCore{
		Core:   c.Core.With(fields),
		module: module,
		levels: c.levels,
	}
}

// Check consults this core's per-module Enabled (NOT the embedded core's) before
// adding itself to the checked entry. This override is required because the
// embedded core's Check would call the embedded Enabled.
func (c *moduleLevelCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}
