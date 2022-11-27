package cli

import (
	"fmt"
	"strconv"
	"time"

	figg "github.com/andydunstall/figg/sdk"
	"github.com/spf13/cobra"
)

type StreamCommand struct {
	Config   *FiggConfig
	cobraCmd *cobra.Command
}

func NewStreamCommand(config *FiggConfig) *StreamCommand {
	command := &StreamCommand{
		Config: config,
	}
	cobraCmd := &cobra.Command{
		Use: "stream",
		RunE: func(cmd *cobra.Command, args []string) error {
			return command.run()
		},
	}
	command.cobraCmd = cobraCmd
	return command
}

func (c *StreamCommand) Run() error {
	return c.cobraCmd.Execute()
}

func (c *StreamCommand) CobraCommand() *cobra.Command {
	return c.cobraCmd
}

func (c *StreamCommand) run() error {
	publisher, err := figg.NewFigg(&figg.Config{
		Addr: c.Config.Addr,
	})
	if err != nil {
		return err
	}

	subscriber, err := figg.NewFigg(&figg.Config{
		Addr: c.Config.Addr,
	})
	if err != nil {
		return err
	}

	last := -1
	subscriber.Subscribe("stream-topic", func(topic string, m []byte) {
		n, _ := strconv.Atoi(string(m))
		if last != -1 && n != last+1 {
			fmt.Println("inconsistent messages")
		}
		last = n

		if n%50 == 0 {
			fmt.Println("received", n)
		}
	})

	for i := 0; ; i++ {
		publisher.Publish("stream-topic", []byte(fmt.Sprintf("%d", i)))
		<-time.After(time.Millisecond * 100)
	}

	return nil
}
