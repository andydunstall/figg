package cli

import (
	"github.com/spf13/cobra"
)

type ClusterCommand struct {
	cobraCmd *cobra.Command
}

func NewClusterCommand() *ClusterCommand {
	command := &ClusterCommand{}
	cobraCmd := &cobra.Command{
		Use: "cluster",
		Run: func(cmd *cobra.Command, args []string) {},
	}
	command.cobraCmd = cobraCmd

	createCommand := NewClusterCreateCommand()
	command.cobraCmd.AddCommand(createCommand.CobraCommand())

	return command
}

func (c *ClusterCommand) Run() error {
	return c.cobraCmd.Execute()
}

func (c *ClusterCommand) CobraCommand() *cobra.Command {
	return c.cobraCmd
}
