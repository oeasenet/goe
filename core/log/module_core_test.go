package log

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// newObservedLogger builds a *zapLogger whose core is a moduleLevelCore wrapping
// an in-memory observer, so tests can assert exactly which entries pass the gate.
func newObservedLogger(global zapcore.Level, overrides map[string]zapcore.Level) (*zapLogger, *observer.ObservedLogs) {
	obs, logs := observer.New(zapcore.DebugLevel)
	core := newModuleLevelCore(obs, &moduleLevels{global: global, overrides: overrides})
	zl := zap.New(core)
	return &zapLogger{logger: zl, sugar: zl.Sugar()}, logs
}

func TestModuleCore_GlobalLevel(t *testing.T) {
	logger, logs := newObservedLogger(zapcore.InfoLevel, nil)
	logger.Debug("dropped")
	logger.Info("kept")
	if logs.Len() != 1 {
		t.Fatalf("got %d entries, want 1", logs.Len())
	}
	if logs.All()[0].Message != "kept" {
		t.Errorf("unexpected message %q", logs.All()[0].Message)
	}
}

func TestModuleCore_OptInDebugPerModule(t *testing.T) {
	logger, logs := newObservedLogger(zapcore.InfoLevel, map[string]zapcore.Level{"job": zapcore.DebugLevel})
	logger.With("module", "job").Debug("job debug") // module opted in -> kept
	logger.With("module", "db").Debug("db debug")   // not opted in -> dropped
	logger.Debug("root debug")                      // root at info -> dropped
	if logs.Len() != 1 {
		t.Fatalf("got %d entries, want 1: %+v", logs.Len(), logs.All())
	}
	if logs.All()[0].Message != "job debug" {
		t.Errorf("unexpected message %q", logs.All()[0].Message)
	}
}

func TestModuleCore_SilenceModuleBelowGlobalDebug(t *testing.T) {
	// Global debug (escape hatch) but gorm dialed down to warn.
	logger, logs := newObservedLogger(zapcore.DebugLevel, map[string]zapcore.Level{"gorm": zapcore.WarnLevel})
	logger.With("module", "gorm").Info("query")  // gorm at warn -> dropped
	logger.With("module", "gorm").Warn("slow")   // gorm at warn -> kept
	logger.With("module", "http").Debug("trace") // other module follows global debug -> kept
	if logs.Len() != 2 {
		t.Fatalf("got %d entries, want 2: %+v", logs.Len(), logs.All())
	}
}

func TestModuleCore_LevelByModuleField(t *testing.T) {
	logger, logs := newObservedLogger(zapcore.InfoLevel, map[string]zapcore.Level{"job": zapcore.DebugLevel})
	// Tag survives across additional .With() calls.
	logger.With("module", "job").With("worker_id", 3).Debug("nested")
	if logs.Len() != 1 {
		t.Fatalf("got %d entries, want 1", logs.Len())
	}
}
