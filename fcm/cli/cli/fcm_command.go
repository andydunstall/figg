package cli

import (
	"github.com/spf13/cobra"
)

var (
	fcmAddr string
)

type FCMCommand struct {
	cobraCmd *cobra.Command
}

func NewFCMCommand() *FCMCommand {
	var cobraCmd = &cobra.Command{
		Use: "fcm",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	return &FCMCommand{
		cobraCmd: cobraCmd,
	}
}

func (c *FCMCommand) Run() error {
	return c.cobraCmd.Execute()
}

func (c *FCMCommand) CobraCommand() *cobra.Command {
	return c.cobraCmd
}

func (c *FCMCommand) AddCommand(command Command) {
	c.cobraCmd.AddCommand(command.CobraCommand())
}
