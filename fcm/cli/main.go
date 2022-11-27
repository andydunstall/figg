package main

import (
	"fmt"
	"os"

	"github.com/andydunstall/figg/fcm/cli/cli"
)

func main() {
	cli := cli.NewCLI()
	if err := cli.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
