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

func New(level ...Level) *Log {
	ll := LevelProd
	if len(level) == 0 {
		ll = LevelDev
	} else {
		ll = level[0]
	}
	j := &Log{}
	zapCfg := zap.NewProductionConfig()
	if ll == LevelDev {
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	j.zapLogger, _ = zapCfg.Build()
	j.zapSugar = j.zapLogger.Sugar().WithOptions(zap.AddCallerSkip(1))
	return j
}
