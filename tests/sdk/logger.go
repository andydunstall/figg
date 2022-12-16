package sdk

import (
	"os"

	"go.uber.org/zap"
)

func setupLogger() *zap.Logger {
	if os.Getenv("VERBOSE") != "" {
		logger, _ := zap.NewDevelopment()
		return logger
	}
	return zap.NewNop()
}
