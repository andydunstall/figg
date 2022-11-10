package cmd

import (
	"fmt"

	figg "github.com/andydunstall/figg/sdk"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(subCmd)
}

var subCmd = &cobra.Command{
	Use:  "sub",
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := figg.NewFigg(&figg.Config{
			Addr: figgAddr,
		})
		if err != nil {
			panic(err)
		}
		sub := figg.NewChannelMessageSubscriber()
		client.Subscribe(args[0], sub)

		for {
			select {
			case m := <-sub.Ch():
				fmt.Println("<-", string(m))
			}
		}
	},
}
