package cli

import (
	"github.com/spf13/cobra"
)

type ChaosCommand struct {
	cobraCmd *cobra.Command
}

func NewChaosCommand() *ChaosCommand {
	command := &ChaosCommand{}
	cobraCmd := &cobra.Command{
		Use: "chaos",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	command.cobraCmd = cobraCmd

	partitionCommand := NewChaosPartitionCommand()
	command.cobraCmd.AddCommand(partitionCommand.CobraCommand())

	return command
}

func (c *ChaosCommand) Run() error {
	return c.cobraCmd.Execute()
}

func (c *ChaosCommand) CobraCommand() *cobra.Command {
	return c.cobraCmd
}
