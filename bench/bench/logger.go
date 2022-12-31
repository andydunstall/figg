package bench

import (
	"go.uber.org/zap"
)

func setupLogger(debugMode bool) *zap.Logger {
	if debugMode {
		logger, _ := zap.NewDevelopment()
		return logger
	}
	return zap.NewNop()
}
