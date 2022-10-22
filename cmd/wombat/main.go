package main

import (
	"log"

	"github.com/andydunstall/wombat/pkg/config"
	"github.com/andydunstall/wombat/pkg/server"
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

func main() {
	config, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("failed to parse config")
	}

	logger, err := setupLogger(config.Verbose)
	if err != nil {
		log.Fatalf("failed to setup logger: %s", err)
	}
	defer logger.Sync()

	logger.Info("starting wombat")

	server := server.NewServer(logger)
	log.Fatal(server.Listen(config.Addr))
}
