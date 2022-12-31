package bench

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	figg "github.com/andydunstall/figg/sdk"
)

func Bench(config *Config) error {
	fmt.Printf("starting benchmark [%s]\n", config)

	wg := &sync.WaitGroup{}

	// Start subscribers before publishing.
	subSamples := make(chan *sample, config.Subscribers)
	for i := 0; i != config.Subscribers; i++ {
		if err := runSubscriber(false, config, wg, subSamples); err != nil {
			return err
		}
	}

	// Divide messages evenly among the publishers.
	pubSamples := make(chan *sample, config.Publishers)
	pubMsgs := MsgsPerClient(config.Messages, config.Publishers)
	for i := 0; i != config.Publishers; i++ {
		if err := runPublisher(pubMsgs[i], config, wg, pubSamples); err != nil {
			return err
		}
	}

	wg.Wait()

	resumeSamples := make(chan *sample, config.Resumers)
	for i := 0; i != config.Resumers; i++ {
		if err := runSubscriber(true, config, wg, resumeSamples); err != nil {
			return err
		}
	}

	wg.Wait()

	if config.Subscribers > 0 {
		subSampleGroup := newSampleGroup()
		close(subSamples)
		for s := range subSamples {
			subSampleGroup.AddSample(s)
		}
		fmt.Println("Sub stats:", subSampleGroup)
	}

	if config.Publishers > 0 {
		pubSampleGroup := newSampleGroup()
		close(pubSamples)
		for s := range pubSamples {
			pubSampleGroup.AddSample(s)
		}
		fmt.Println("Pub stats:", pubSampleGroup)
	}

	if config.Resumers > 0 {
		resumeSampleGroup := newSampleGroup()
		close(resumeSamples)
		for s := range resumeSamples {
			resumeSampleGroup.AddSample(s)
		}
		fmt.Println("Resume stats:", resumeSampleGroup)
	}

	return nil
}

func runPublisher(messages int, config *Config, wg *sync.WaitGroup, pubSamples chan *sample) error {
	conn, err := figg.Connect(
		config.Addr,
		figg.WithLogger(setupLogger(config.Verbose)),
	)
	if err != nil {
		return err
	}

	// ch receives the start time and end time from the subscriber.
	ch := make(chan time.Time, 2)

	message := make([]byte, config.MessageSize)
	rand.Read(message)

	acked := 0
	for i := 0; i != messages; i++ {
		conn.Publish(config.Topic, message, func() {
			acked++

			if acked == 1 {
				ch <- time.Now()
			}
			if acked == messages {
				ch <- time.Now()
			}
		})
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer conn.Close()

		start := <-ch
		end := <-ch

		pubSamples <- newSample(messages, config.MessageSize, start, end)
	}()

	return nil
}

// runSubscriber subscribes to the configured topic and spawns a goroutine
// so wait for all messages to be received.
func runSubscriber(resume bool, config *Config, wg *sync.WaitGroup, subSamples chan *sample) error {
	conn, err := figg.Connect(
		config.Addr,
		figg.WithLogger(setupLogger(config.Verbose)),
	)
	if err != nil {
		return err
	}

	// ch receives the start time and end time from the subscriber.
	ch := make(chan time.Time, 2)

	received := 0
	// If resuming start from an offset of 0.
	opts := []figg.TopicOption{}
	if resume {
		opts = append(opts, figg.WithOffset(0))
	}
	conn.Subscribe(config.Topic, func(m *figg.Message) {
		received++

		if received == 1 {
			ch <- time.Now()
		}
		if received == config.Messages {
			ch <- time.Now()
		}
	}, opts...)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer conn.Close()

		start := <-ch
		end := <-ch

		subSamples <- newSample(config.Messages, config.MessageSize, start, end)
	}()

	return nil
}

// MsgsPerClient divides the number of messages by the number of clients and tries to distribute them as evenly as possible
func MsgsPerClient(numMsgs, numClients int) []int {
	var counts []int
	if numClients == 0 || numMsgs == 0 {
		return counts
	}
	counts = make([]int, numClients)
	mc := numMsgs / numClients
	for i := 0; i < numClients; i++ {
		counts[i] = mc
	}
	extra := numMsgs % numClients
	for i := 0; i < extra; i++ {
		counts[i]++
	}
	return counts
}
