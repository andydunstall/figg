package cli

import (
	"time"

	figg "github.com/andydunstall/figg/sdk"
	"github.com/spf13/cobra"
)

type PublishCommand struct {
	Config   *FiggConfig
	cobraCmd *cobra.Command
}

func NewPublishCommand(config *FiggConfig) *PublishCommand {
	command := &PublishCommand{
		Config: config,
	}
	cobraCmd := &cobra.Command{
		Use:  "pub",
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			topic := args[0]
			message := []byte(args[1])
			return command.run(topic, message)
		},
	}
	command.cobraCmd = cobraCmd
	return command
}

func (c *PublishCommand) Run() error {
	return c.cobraCmd.Execute()
}

func (c *PublishCommand) CobraCommand() *cobra.Command {
	return c.cobraCmd
}

func (c *PublishCommand) run(topic string, message []byte) error {
	client, err := figg.NewFigg(&figg.Config{
		Addr: c.Config.Addr,
	})
	if err != nil {
		return err
	}
	client.Publish(topic, message)
	// Given time to publish.
	// TODO(AD) Publish should block until ACKed.
	<-time.After(time.Second)
	return nil
}
