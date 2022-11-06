package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/andydunstall/figg/service"
	"github.com/andydunstall/figg/service/pkg/config"
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
		log.Fatalf("failed to parse config: %s", err)
	}

	logger, err := setupLogger(config.Verbose)
	if err != nil {
		log.Fatalf("failed to setup logger: %s", err)
	}
	defer logger.Sync()

	doneCh := make(chan interface{})
	service.Run(config, logger, doneCh)

	// TODO(AD) this isn't doing anything yet as Run blocks
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	// Block until we receive our signal.
	<-c
	close(doneCh)

	logger.Info("received interrupt")
}
