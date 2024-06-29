package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type log struct {
	zapLogger *zap.Logger
	zapSugar  *zap.SugaredLogger
}

func (jl *log) Debug(args ...any) {
	jl.zapSugar.Debug(args...)
}

func (jl *log) Log(args ...any) {
	jl.zapSugar.Log(zapcore.InfoLevel, args...)
}

func (jl *log) Info(args ...any) {
	jl.zapSugar.Info(args...)
}

func (jl *log) Warn(args ...any) {
	jl.zapSugar.Warn(args...)
}

func (jl *log) Error(args ...any) {
	jl.zapSugar.Error(args...)
}

func (jl *log) Fatal(args ...any) {
	jl.zapSugar.Fatal(args...)
}

func (jl *log) Panic(args ...any) {
	jl.zapSugar.Panic(args...)
}

func (jl *log) Debugf(format string, args ...any) {
	jl.zapSugar.Debugf(format, args...)
}

func (jl *log) Logf(format string, args ...any) {
	jl.zapSugar.Logf(zapcore.InfoLevel, format, args...)
}

func (jl *log) Infof(format string, args ...any) {
	jl.zapSugar.Infof(format, args...)
}

func (jl *log) Warnf(format string, args ...any) {
	jl.zapSugar.Warnf(format, args...)
}

func (jl *log) Errorf(format string, args ...any) {
	jl.zapSugar.Errorf(format, args...)
}

func (jl *log) Fatalf(format string, args ...any) {
	jl.zapSugar.Fatalf(format, args...)
}

func (jl *log) Panicf(format string, args ...any) {
	jl.zapSugar.Panicf(format, args...)
}

func (jl *log) Debugw(msg string, keysAndValues ...any) {
	jl.zapSugar.Debugw(msg, keysAndValues...)
}

func (jl *log) Infow(msg string, keysAndValues ...any) {
	jl.zapSugar.Infow(msg, keysAndValues...)
}

func (jl *log) Warnw(msg string, keysAndValues ...any) {
	jl.zapSugar.Warnw(msg, keysAndValues...)
}

func (jl *log) Errorw(msg string, keysAndValues ...any) {
	jl.zapSugar.Errorw(msg, keysAndValues...)
}

func (jl *log) Fatalw(msg string, keysAndValues ...any) {
	jl.zapSugar.Fatalw(msg, keysAndValues...)
}

func (jl *log) Panicw(msg string, keysAndValues ...any) {
	jl.zapSugar.Panicw(msg, keysAndValues...)
}

func (jl *log) GetZapLogger() *zap.Logger {
	return jl.zapLogger
}

func (jl *log) GetZapSugarLogger() *zap.SugaredLogger {
	return jl.zapSugar
}

func (jl *log) Close() {
	jl.zapLogger.Sync()
}
