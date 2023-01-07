package main

import (
	"log"
	"os"
	"os/signal"

	adminService "github.com/andydunstall/figg/server/pkg/admin/service"
	"github.com/andydunstall/figg/server/pkg/config"
	messagingService "github.com/andydunstall/figg/server/pkg/messaging/service"
	"go.uber.org/zap"
)

func setupLogger(debugMode bool) (*zap.Logger, error) {
	if debugMode {
		logger, err := zap.NewDevelopment()
		if err != nil {
			return nil, err
		}
		return logger, nil
	}
	return zap.NewProduction()
}

func waitForInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func main() {
	config, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("failed to parse config: %s", err)
	}

	logger, err := setupLogger(config.Verbose)
	if err != nil {
		log.Fatalf("failed to setup logger: %s", err)
	}
	defer logger.Sync()

	logger.Info("starting figg server", zap.Object("config", config))

	messagingService := messagingService.NewMessagingService(config, logger)
	_, err = messagingService.Serve()
	if err != nil {
		logger.Fatal("failed to start messaging service", zap.Error(err))
	}
	defer messagingService.Close()

	adminService := adminService.NewAdminService(config, logger)
	_, err = adminService.Serve()
	if err != nil {
		logger.Fatal("failed to start admin service", zap.Error(err))
	}
	defer adminService.Close()

	waitForInterrupt()
	logger.Info("received interrupt; exiting")
}
