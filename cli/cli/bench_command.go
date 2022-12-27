package cli

import (
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
	publisher, err := c.connectedClient()
	if err != nil {
		return err
	}
	defer publisher.Close()

	message := make([]byte, payloadLen)
	rand.Read(message)

	start := time.Now()

	for i := 0; i != c.publishes; i++ {
		publisher.PublishWaitForACK("bench-publish", message)
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
	publisher, err := c.connectedClient()
	if err != nil {
		return err
	}
	defer publisher.Close()

	subscriber, err := c.connectedClient()
	if err != nil {
		return err
	}
	defer subscriber.Close()

	message := make([]byte, payloadLen)
	rand.Read(message)

	doneCh := make(chan interface{})

	count := c.publishes
	received := 0
	subscriber.Subscribe("bench-subscribe", func(m *figg.Message) {
		received++
		if received == count {
			close(doneCh)
		}
	})

	start := time.Now()

	for i := 0; i != count; i++ {
		publisher.PublishWaitForACK("bench-subscribe", message)
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
	publisher, err := c.connectedClient()
	if err != nil {
		return err
	}
	defer publisher.Close()

	message := make([]byte, payloadLen)
	rand.Read(message)
	for i := 0; i != c.publishes; i++ {
		publisher.PublishWaitForACK("bench-resume", message)
	}

	subscriber, err := c.connectedClient()
	if err != nil {
		return err
	}
	defer subscriber.Close()

	doneCh := make(chan interface{})

	start := time.Now()

	count := c.publishes
	received := 0
	subscriber.Subscribe("bench-resume", func(m *figg.Message) {
		received++
		if received == count {
			close(doneCh)
		}
	}, figg.WithOffset(0))

	<-doneCh

	elapsed := time.Since(start)
	fmt.Printf("  ====== SAMPLE %d ======\n", i)
	fmt.Printf("  requests: %d\n", c.publishes)
	fmt.Printf("  payload size: %d\n", payloadLen)
	fmt.Printf("  elapsed: %s\n", elapsed)
	fmt.Println("")

	return nil
}

func (c *BenchCommand) connectedClient() (*figg.Figg, error) {
	client, err := figg.Connect(
		c.Config.Addr,
		figg.WithLogger(setupLogger(c.Config.Verbose)),
	)
	if err != nil {
		return nil, err
	}
	return client, nil
}
