package mongodb

import (
	"fmt"
	"log/slog"
)

type Logger interface {
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Panic(args ...any)

	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Panicf(format string, args ...any)
}

type defaultLogger struct {
	defaultSloger *slog.Logger
}

func newDefaultLogger() Logger {
	return &defaultLogger{
		defaultSloger: slog.Default(),
	}
}

func (d *defaultLogger) Debug(args ...any) {
	d.defaultSloger.Debug(args[0].(string), args[1:]...)
}

func (d *defaultLogger) Info(args ...any) {
	d.defaultSloger.Info(args[0].(string), args[1:]...)
}

func (d *defaultLogger) Warn(args ...any) {
	d.defaultSloger.Warn(args[0].(string), args[1:]...)
}

func (d *defaultLogger) Error(args ...any) {
	d.defaultSloger.Error(args[0].(string), args[1:]...)
}

func (d *defaultLogger) Panic(args ...any) {
	d.defaultSloger.Error(args[0].(string), args[1:]...)
	panic(args[0])
}

func (d *defaultLogger) Debugf(format string, args ...any) {
	d.Debug(fmt.Sprintf(format, args...))
}

func (d *defaultLogger) Infof(format string, args ...any) {
	d.Info(fmt.Sprintf(format, args...))
}

func (d *defaultLogger) Warnf(format string, args ...any) {
	d.Warn(fmt.Sprintf(format, args...))
}

func (d *defaultLogger) Errorf(format string, args ...any) {
	d.Error(fmt.Sprintf(format, args...))
}

func (d *defaultLogger) Panicf(format string, args ...any) {
	d.Panic(fmt.Sprintf(format, args...))
}
