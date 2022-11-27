package cli

import (
	fcm "github.com/andydunstall/figg/fcm/sdk"
	"github.com/spf13/cobra"
)

type ChaosPartitionCommand struct {
	cobraCmd    *cobra.Command
	node        string
	chaosConfig fcm.ChaosConfig
}

func NewChaosPartitionCommand() *ChaosPartitionCommand {
	command := &ChaosPartitionCommand{}
	cobraCmd := &cobra.Command{
		Use: "partition",
		RunE: func(cmd *cobra.Command, args []string) error {
			return command.run()
		},
	}
	command.cobraCmd = cobraCmd

	command.cobraCmd.PersistentFlags().StringVar(&command.node, "node", "", "")
	command.cobraCmd.MarkPersistentFlagRequired("node")

	command.cobraCmd.PersistentFlags().IntVar(&command.chaosConfig.Repeat, "repeat", 0, "")
	command.cobraCmd.PersistentFlags().IntVar(&command.chaosConfig.Duration, "duration", 0, "")

	return command
}

func (c *ChaosPartitionCommand) Run() error {
	return c.cobraCmd.Execute()
}

func (c *ChaosPartitionCommand) CobraCommand() *cobra.Command {
	return c.cobraCmd
}

func (c *ChaosPartitionCommand) run() error {
	fcm := fcm.NewFCM()

	if err := fcm.AddChaosPartition(c.node, c.chaosConfig); err != nil {
		return err
	}

	// cluster, err := fcm.AddCluster()
	// if err != nil {
	// 	return err
	// }

	// fmt.Println("")
	// fmt.Println("    Cluster")
	// fmt.Println("    -------")
	// fmt.Println("    ID:", cluster.ID)
	// fmt.Println("")
	// fmt.Println("    Nodes")
	// fmt.Println("    -------")
	// for _, node := range cluster.Nodes {
	// 	fmt.Println("    ID: ", node.ID, "|", "Addr:", node.Addr, "|", "Proxy Addr:", node.ProxyAddr)
	// }
	// fmt.Println("")

	return nil
}
