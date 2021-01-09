package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(verbose, debug, logJSON bool) (*zap.Logger, error) {
	var config zap.Config

	if logJSON {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
	}

	// default configuration for all encoders
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	config.Development = false
	config.DisableStacktrace = true
	config.DisableCaller = true

	if debug {
		config.Development = true
		config.DisableStacktrace = false
		config.DisableCaller = false
	}

	if verbose || debug {
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	return config.Build()
}
