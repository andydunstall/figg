package main

import (
	"log"

	"github.com/andydunstall/wombat/service/pkg/config"
	"github.com/andydunstall/wombat/service/pkg/wombat"
)

func main() {
	config, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("failed to parse config: %s", err)
	}

	wombat.Run(config)
}
