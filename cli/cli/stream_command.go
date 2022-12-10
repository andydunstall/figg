package cli

import (
	"context"
	"fmt"
	"strconv"
	"time"

	figg "github.com/andydunstall/figg/sdk"
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
	pubStateSubscriber := figg.NewChannelStateSubscriber()
	publisher, err := figg.NewFigg(&figg.Config{
		Addr:            c.PubAddr,
		StateSubscriber: pubStateSubscriber,
	})
	if err != nil {
		return err
	}

	subStateSubscriber := figg.NewChannelStateSubscriber()
	subscriber, err := figg.NewFigg(&figg.Config{
		Addr:            c.SubAddr,
		StateSubscriber: subStateSubscriber,
	})
	if err != nil {
		return err
	}

	last := 0
	subscriber.Subscribe(context.Background(), "stream-topic", func(topic string, m []byte) {
		n, _ := strconv.Atoi(string(m))
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
			publisher.Publish(context.Background(), "stream-topic", []byte(fmt.Sprintf("%d", i)))
			i++
		case state := <-subStateSubscriber.Ch():
			fmt.Println("sub state", figg.StateToString(state))
		case state := <-pubStateSubscriber.Ch():
			fmt.Println("pub state", figg.StateToString(state))
		}
	}

	return nil
}
