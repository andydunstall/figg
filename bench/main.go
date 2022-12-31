package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/andydunstall/figg/bench/bench"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	config, err := bench.ParseConfig()
	if err != nil {
		fmt.Println("failed to parse config")
		os.Exit(1)
	}

	if err := bench.Bench(config); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
