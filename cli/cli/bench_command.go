package cli

import (
	"fmt"
	"time"

	figg "github.com/andydunstall/figg/sdk"
	"github.com/spf13/cobra"
)

type BenchCommand struct {
	Config    *FiggConfig
	cobraCmd  *cobra.Command
	samples   int
	publishes int
}

func NewBenchCommand(config *FiggConfig) *BenchCommand {
	command := &BenchCommand{
		Config: config,
	}
	cobraCmd := &cobra.Command{
		Use: "bench",
		RunE: func(cmd *cobra.Command, args []string) error {
			return command.run()
		},
	}
	cobraCmd.PersistentFlags().IntVar(&command.samples, "samples", 1, "number of bench samples")
	cobraCmd.PersistentFlags().IntVar(&command.publishes, "publishes", 10000, "number of publishes")
	command.cobraCmd = cobraCmd
	return command
}

func (c *BenchCommand) Run() error {
	return c.cobraCmd.Execute()
}

func (c *BenchCommand) CobraCommand() *cobra.Command {
	return c.cobraCmd
}

func (c *BenchCommand) run() error {
	for i := 0; i != c.samples; i++ {
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

		<-time.After(time.Second)

		// Wait to become connected and attached.
		// TODO(AD) this should be an event (maybe Subscribe should block until
		// received ATTACHED)

		doneCh := make(chan interface{})

		count := c.publishes
		received := 0
		subscriber.Subscribe("bench-topic", func(topic string, m []byte) {
			received++
			if received == count {
				close(doneCh)
			}
		})

		<-time.After(time.Second)

		start := time.Now()

		for i := 0; i != count; i++ {
			publisher.Publish("bench-topic", []byte(fmt.Sprintf("message-%d", i)))
		}

		<-doneCh

		elapsed := time.Since(start)
		fmt.Println("elapsed", elapsed, "count", count)
	}
	return nil
}
