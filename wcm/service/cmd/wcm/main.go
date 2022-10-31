package main

import (
	"github.com/andydunstall/wombat/wcm/service/pkg/server"
	"go.uber.org/zap"
)

const (
	Addr = "127.0.0.1:7229"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	logger.Info("starting wcm service")

	server := server.NewServer(logger)
	if err := server.Listen(Addr); err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
}
