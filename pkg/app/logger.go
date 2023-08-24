package app

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func getLogger(name string) *zap.Logger {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.RFC3339TimeEncoder

	zapCfg := zap.NewProductionConfig()
	zapCfg.EncoderConfig = encoderCfg

	logger := zap.Must(zapCfg.Build())

	return logger.Named(name)
}
