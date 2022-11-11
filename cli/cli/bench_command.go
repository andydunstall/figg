package cli

import (
	"fmt"
	"sync"
	"time"

	figg "github.com/andydunstall/figg/sdk"
	"github.com/spf13/cobra"
)

type BenchCommand struct {
	Config   *FiggConfig
	cobraCmd *cobra.Command
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
	sub := figg.NewChannelMessageSubscriber()
	subscriber.Subscribe("bench-topic", sub)

	<-time.After(time.Second)

	start := time.Now()

	count := 10000

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		received := 0
		for {
			select {
			case <-sub.Ch():
				received += 1
				if received == count {
					return
				}
			}
		}
	}()

	for i := 0; i != count; i++ {
		publisher.Publish("bench-topic", []byte(fmt.Sprintf("message-%d", i)))
	}

	wg.Wait()

	elapsed := time.Since(start)
	fmt.Println("elapsed", elapsed, "count", count)

	return nil
}
