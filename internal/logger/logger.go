package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Mode string

const (
	ModeDevelopment Mode = "development"
	ModeProduction  Mode = "production"
)

func NewLogger(mode Mode) (*zap.Logger, *zap.SugaredLogger) {
	var cfg zap.Config

	switch mode {
	case ModeProduction:
		cfg = zap.NewProductionConfig()
	case ModeDevelopment:
		fallthrough
	default:
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	logger, err := cfg.Build()
	if err != nil {
		fallback, _ := zap.NewDevelopment()
		fallback.Fatal("failed to initialize zap logger", zap.Error(err))
	}

	return logger, logger.Sugar()
}
