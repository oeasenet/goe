package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Log struct {
	zapLogger *zap.Logger
	zapSugar  *zap.SugaredLogger
}

func (jl *Log) Debug(args ...any) {
	jl.zapSugar.Debug(args...)
}

func (jl *Log) Log(args ...any) {
	jl.zapSugar.Log(zapcore.InfoLevel, args...)
}

func (jl *Log) Info(args ...any) {
	jl.zapSugar.Info(args...)
}

func (jl *Log) Warn(args ...any) {
	jl.zapSugar.Warn(args...)
}

func (jl *Log) Error(args ...any) {
	jl.zapSugar.Error(args...)
}

func (jl *Log) Fatal(args ...any) {
	jl.zapSugar.Fatal(args...)
}

func (jl *Log) Panic(args ...any) {
	jl.zapSugar.Panic(args...)
}

func (jl *Log) Debugf(format string, args ...any) {
	jl.zapSugar.Debugf(format, args...)
}

func (jl *Log) Logf(format string, args ...any) {
	jl.zapSugar.Logf(zapcore.InfoLevel, format, args...)
}

func (jl *Log) Infof(format string, args ...any) {
	jl.zapSugar.Infof(format, args...)
}

func (jl *Log) Warnf(format string, args ...any) {
	jl.zapSugar.Warnf(format, args...)
}

func (jl *Log) Errorf(format string, args ...any) {
	jl.zapSugar.Errorf(format, args...)
}

func (jl *Log) Fatalf(format string, args ...any) {
	jl.zapSugar.Fatalf(format, args...)
}

func (jl *Log) Panicf(format string, args ...any) {
	jl.zapSugar.Panicf(format, args...)
}

func (jl *Log) Debugw(msg string, keysAndValues ...any) {
	jl.zapSugar.Debugw(msg, keysAndValues...)
}

func (jl *Log) Infow(msg string, keysAndValues ...any) {
	jl.zapSugar.Infow(msg, keysAndValues...)
}

func (jl *Log) Warnw(msg string, keysAndValues ...any) {
	jl.zapSugar.Warnw(msg, keysAndValues...)
}

func (jl *Log) Errorw(msg string, keysAndValues ...any) {
	jl.zapSugar.Errorw(msg, keysAndValues...)
}

func (jl *Log) Fatalw(msg string, keysAndValues ...any) {
	jl.zapSugar.Fatalw(msg, keysAndValues...)
}

func (jl *Log) Panicw(msg string, keysAndValues ...any) {
	jl.zapSugar.Panicw(msg, keysAndValues...)
}

func (jl *Log) GetZapLogger() *zap.Logger {
	return jl.zapLogger
}

func (jl *Log) GetZapSugarLogger() *zap.SugaredLogger {
	return jl.zapSugar
}

func (jl *Log) Close() {
	jl.zapLogger.Sync()
}
