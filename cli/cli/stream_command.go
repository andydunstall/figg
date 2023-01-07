package cli

import (
	"fmt"
	"strconv"
	"time"

	figg "github.com/andydunstall/figg/sdk/go"
	"github.com/spf13/cobra"
)

type StreamCommand struct {
	Config   *FiggConfig
	PubAddr  string
	SubAddr  string
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

	cobraCmd.PersistentFlags().StringVar(&command.PubAddr, "pub-addr", "127.0.0.1:8119", "publisher figg cluster address")
	cobraCmd.PersistentFlags().StringVar(&command.SubAddr, "sub-addr", "127.0.0.1:8119", "subscriber figg cluster address")

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
	publisher, err := figg.Connect(c.PubAddr, figg.WithConnStateChangeCB(func(s figg.ConnState) {
		fmt.Println("pub state", s)
	}))
	if err != nil {
		return err
	}

	subscriber, err := figg.Connect(c.SubAddr, figg.WithConnStateChangeCB(func(s figg.ConnState) {
		fmt.Println("sub state", s)
	}))
	if err != nil {
		return err
	}

	last := 0
	subscriber.Subscribe("stream-topic", func(m *figg.Message) {
		n, _ := strconv.Atoi(string(m.Data))
		if last != 0 && n != last+1 {
			panic("out of order messages")
		}
		last = n

		if n%50 == 0 {
			fmt.Println("received", n)
		}
	})

	ticker := time.NewTicker(10 * time.Millisecond)
	i := 1
	for {
		select {
		case <-ticker.C:
			publisher.PublishWaitForACK("stream-topic", []byte(fmt.Sprintf("%d", i)))
			i++
		}
	}

	return nil
}
