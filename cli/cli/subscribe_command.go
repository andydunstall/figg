package cli

import (
	"context"
	"fmt"

	figg "github.com/andydunstall/figg/sdk"
	"github.com/spf13/cobra"
)

type SubscribeCommand struct {
	Config   *FiggConfig
	cobraCmd *cobra.Command
}

func NewSubscribeCommand(config *FiggConfig) *SubscribeCommand {
	command := &SubscribeCommand{
		Config: config,
	}
	cobraCmd := &cobra.Command{
		Use:  "sub",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			topic := args[0]
			return command.run(topic)
		},
	}
	command.cobraCmd = cobraCmd
	return command
}

func (c *SubscribeCommand) Run() error {
	return c.cobraCmd.Execute()
}

func (c *SubscribeCommand) CobraCommand() *cobra.Command {
	return c.cobraCmd
}

func (c *SubscribeCommand) run(topic string) error {
	client, err := figg.NewFigg(&figg.Config{
		Addr: c.Config.Addr,
	})
	if err != nil {
		return err
	}
	client.Subscribe(context.Background(), topic, func(topic string, m []byte) {
		fmt.Println("<-", string(m))
	})

	select {}
	return nil
}
