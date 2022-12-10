package cli

import (
	"context"
	"fmt"
	"math/rand"
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
	cobraCmd.PersistentFlags().IntVar(&command.samples, "samples", 5, "number of bench samples")
	cobraCmd.PersistentFlags().IntVar(&command.publishes, "publishes", 100000, "number of publishes")
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
	if err := c.runPublish(); err != nil {
		return err
	}
	if err := c.runSubscribe(); err != nil {
		return err
	}
	if err := c.runResume(); err != nil {
		return err
	}
	return nil
}

func (c *BenchCommand) runPublish() error {
	fmt.Println("====== PUBLISH ======")
	for i := 0; i != c.samples; i++ {
		if err := c.samplePublish(i, 1024); err != nil {
			return err
		}
	}
	return nil
}

func (c *BenchCommand) samplePublish(i int, payloadLen int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	publisher, err := c.connectedClient(ctx)
	if err != nil {
		return err
	}
	defer publisher.Shutdown()

	message := make([]byte, payloadLen)
	rand.Read(message)

	start := time.Now()

	// Only wait for the last ACK (since this means all previous messages have
	// also been ACK'ed).
	for i := 0; i != c.publishes-1; i++ {
		publisher.PublishNoACK("bench-publish", message)
	}
	if err := publisher.Publish(context.Background(), "bench-publish", message); err != nil {
		return err
	}

	elapsed := time.Since(start)
	fmt.Printf("  ====== SAMPLE %d ======\n", i)
	fmt.Printf("  requests: %d\n", c.publishes)
	fmt.Printf("  payload size: %d\n", payloadLen)
	fmt.Printf("  elapsed: %s\n", elapsed)
	fmt.Println("")

	return nil
}

func (c *BenchCommand) runSubscribe() error {
	fmt.Println("====== SUBSCRIBE ======")
	for i := 0; i != c.samples; i++ {
		if err := c.sampleSubscribe(i, 1024); err != nil {
			return err
		}
	}
	return nil
}

func (c *BenchCommand) sampleSubscribe(i int, payloadLen int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	publisher, err := c.connectedClient(ctx)
	if err != nil {
		return err
	}
	defer publisher.Shutdown()

	subscriber, err := c.connectedClient(ctx)
	if err != nil {
		return err
	}
	defer subscriber.Shutdown()

	message := make([]byte, payloadLen)
	rand.Read(message)

	doneCh := make(chan interface{})

	count := c.publishes
	received := 0
	subscriber.Subscribe("bench-subscribe", func(topic string, m []byte) {
		received++
		if received == count {
			close(doneCh)
		}
	})

	start := time.Now()

	for i := 0; i != count; i++ {
		publisher.PublishNoACK("bench-subscribe", message)
	}

	<-doneCh

	elapsed := time.Since(start)
	fmt.Printf("  ====== SAMPLE %d ======\n", i)
	fmt.Printf("  requests: %d\n", c.publishes)
	fmt.Printf("  payload size: %d\n", payloadLen)
	fmt.Printf("  elapsed: %s\n", elapsed)
	fmt.Println("")

	return nil
}

func (c *BenchCommand) runResume() error {
	fmt.Println("====== RESUME ======")
	for i := 0; i != c.samples; i++ {
		if err := c.sampleResume(i, 1024); err != nil {
			return err
		}
	}
	return nil
}

func (c *BenchCommand) sampleResume(i int, payloadLen int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	publisher, err := c.connectedClient(ctx)
	if err != nil {
		return err
	}
	defer publisher.Shutdown()

	message := make([]byte, payloadLen)
	rand.Read(message)
	for i := 0; i != c.publishes-1; i++ {
		publisher.PublishNoACK("bench-resume", message)
	}
	// Wait for the final publish to be ACK'ed (which means all others have
	// been ACK'ed).
	if err = publisher.Publish(context.Background(), "bench-resume", message); err != nil {
		return err
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	subscriber, err := c.connectedClient(ctx)
	if err != nil {
		return err
	}
	defer subscriber.Shutdown()

	doneCh := make(chan interface{})

	start := time.Now()

	count := c.publishes
	received := 0
	subscriber.SubscribeFromOffset("bench-resume", "0", func(topic string, m []byte) {
		received++
		if received == count {
			close(doneCh)
		}
	})

	<-doneCh

	elapsed := time.Since(start)
	fmt.Printf("  ====== SAMPLE %d ======\n", i)
	fmt.Printf("  requests: %d\n", c.publishes)
	fmt.Printf("  payload size: %d\n", payloadLen)
	fmt.Printf("  elapsed: %s\n", elapsed)
	fmt.Println("")

	return nil
}

// connectedClient returns a client after waiting for it to connect.
func (c *BenchCommand) connectedClient(ctx context.Context) (*figg.Figg, error) {
	stateSub := figg.NewChannelStateSubscriber()
	client, err := figg.NewFigg(&figg.Config{
		Addr:            c.Config.Addr,
		StateSubscriber: stateSub,
	})
	if err != nil {
		return nil, err
	}

	if err = stateSub.WaitForConnected(ctx); err != nil {
		return nil, err
	}

	return client, nil
}
