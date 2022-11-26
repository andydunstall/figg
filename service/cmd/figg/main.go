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

func setupCPUProfile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	pprof.StartCPUProfile(f)
	return nil
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

	if config.CPUProfile != "" {
		if err := setupCPUProfile(config.CPUProfile); err != nil {
			logger.Fatal("failed to start cpu profile", zap.Error(err))
		}
		logger.Info("started cpu profile", zap.String("output", config.CPUProfile))
		defer pprof.StopCPUProfile()
	}

	doneCh := make(chan interface{})
	go func() {
		service.Run(config, logger, doneCh)
	}()

	waitForInterrupt()
	logger.Info("received interrupt")

	close(doneCh)
}
