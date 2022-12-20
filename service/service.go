package service

import (
	"net"

	"github.com/andydunstall/figg/service/pkg/config"
	"github.com/andydunstall/figg/service/pkg/server"
	"github.com/andydunstall/figg/service/pkg/topic"
	"go.uber.org/zap"
)

func Run(config config.Config, logger *zap.Logger, doneCh <-chan interface{}) {
	logger.Info("starting figg service", zap.Object("config", config))

	server := server.NewServer(topic.NewBroker(config.DataDir), logger)

	lis, err := net.Listen("tcp", config.Addr)
	if err != nil {
		logger.Fatal("failed to start listener", zap.Error(err))
	}

	go func() {
		if err := server.Serve(lis); err != nil {
			logger.Error("serve failed", zap.Error(err))
		}
	}()

	<-doneCh

	logger.Info("shutting down")
	lis.Close()
}
