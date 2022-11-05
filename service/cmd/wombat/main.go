package main

import (
	"log"

	"github.com/andydunstall/wombat/service"
	"github.com/andydunstall/wombat/service/pkg/config"
)

func main() {
	config, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("failed to parse config: %s", err)
	}

	service.Run(config)
}
