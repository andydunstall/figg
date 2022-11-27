package cli

import (
	"fmt"

	fcm "github.com/andydunstall/figg/fcm/sdk"
	"github.com/spf13/cobra"
)

type ClusterCreateCommand struct {
	cobraCmd *cobra.Command
}

func NewClusterCreateCommand() *ClusterCreateCommand {
	command := &ClusterCreateCommand{}
	cobraCmd := &cobra.Command{
		Use: "create",
		RunE: func(cmd *cobra.Command, args []string) error {
			return command.run()
		},
	}
	command.cobraCmd = cobraCmd
	return command
}

func (c *ClusterCreateCommand) Run() error {
	return c.cobraCmd.Execute()
}

func (c *ClusterCreateCommand) CobraCommand() *cobra.Command {
	return c.cobraCmd
}

func (c *ClusterCreateCommand) run() error {
	fcm := fcm.NewFCM()
	cluster, err := fcm.AddCluster()
	if err != nil {
		return err
	}

	fmt.Println("")
	fmt.Println("    Cluster")
	fmt.Println("    -------")
	fmt.Println("    ID:", cluster.ID)
	fmt.Println("")
	fmt.Println("    Nodes")
	fmt.Println("    -------")
	for _, node := range cluster.Nodes {
		fmt.Println("    ID: ", node.ID, "|", "Addr:", node.Addr, "|", "Proxy Addr:", node.ProxyAddr)
	}
	fmt.Println("")

	return nil
}
