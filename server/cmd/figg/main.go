package main

import (
	"log"
	"os"
	"os/signal"
	"runtime/pprof"

	"github.com/andydunstall/figg/server"
	"github.com/andydunstall/figg/server/pkg/config"
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

	if config.MemoryProfile != "" {
		logger.Info("started memory profile", zap.String("output", config.MemoryProfile))
		defer func() {
			f, err := os.Create(config.MemoryProfile)
			if err != nil {
				logger.Fatal("failed to open memory profile", zap.Error(err))
				return
			}
			logger.Info("writing memory profile", zap.String("output", config.MemoryProfile))
			pprof.WriteHeapProfile(f)
		}()
	}

	figg := server.NewFigg(config, logger)
	defer figg.Close()

	go figg.Serve()

	waitForInterrupt()
	logger.Info("received interrupt; exiting")
}
