package main

import (
	"log"
	"os"
	"os/signal"
	"runtime/pprof"

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

	if config.CPUProfile != "" {
		f, err := os.Create(config.CPUProfile)
		if err != nil {
			logger.Error("failed to start cpu profile", zap.Error(err))
		}
		logger.Info("starting cpu profile", zap.String("output", config.CPUProfile))
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	doneCh := make(chan interface{})
	go func() {
		service.Run(config, logger, doneCh)
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	// Block until we receive our signal.
	<-c
	close(doneCh)

	logger.Info("received interrupt")
}
