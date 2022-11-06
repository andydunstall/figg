package main

import (
	"github.com/andydunstall/figg/fcm/service/pkg/server"
	"go.uber.org/zap"
	"net"
)

const (
	Addr = "127.0.0.1:7229"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	logger.Info("starting fcm service")

	lis, err := net.Listen("tcp", Addr)
	if err != nil {
		logger.Fatal("failed to start listener", zap.Error(err))
	}

	server := server.NewServer(logger)
	if err := server.Serve(lis); err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
}
