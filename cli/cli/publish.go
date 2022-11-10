package cmd

import (
	"time"

	figg "github.com/andydunstall/figg/sdk"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pubCmd)
}

var pubCmd = &cobra.Command{
	Use:  "pub",
	Args: cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client, _ := figg.NewFigg(&figg.Config{
			Addr: figgAddr,
		})
		client.Publish(args[0], []byte(args[1]))
		// Given time to publish.
		// TODO(AD) Publish should block until ACKed.
		<-time.After(time.Second)
	},
}
