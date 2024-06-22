package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Level int

const (
	LevelProd Level = iota
	LevelDev
)

// Logger represents a logging interface.
type Logger interface {
	Debug(args ...any)
	Log(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Fatal(args ...any)
	Panic(args ...any)

	Debugf(format string, args ...any)
	Logf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	Panicf(format string, args ...any)

	Debugw(msg string, keysAndValues ...any)
	Infow(msg string, keysAndValues ...any)
	Warnw(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
	Fatalw(msg string, keysAndValues ...any)
	Panicw(msg string, keysAndValues ...any)

	GetZapLogger() *zap.Logger
	GetZapSugarLogger() *zap.SugaredLogger
	Close()
}

func New(level ...Level) Logger {
	ll := LevelProd
	if len(level) == 0 {
		ll = LevelDev
	} else {
		ll = level[0]
	}
	j := &log{}
	zapCfg := zap.NewProductionConfig()
	if ll == LevelDev {
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	j.zapLogger, _ = zapCfg.Build()
	j.zapSugar = j.zapLogger.Sugar().WithOptions(zap.AddCallerSkip(2))
	return j
}
