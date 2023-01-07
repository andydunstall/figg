package figg

import (
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	ErrAlreadySubscribed = errors.New("already subscribed")
)

type Figg struct {
	opts *Options
	conn *connection

	// shutdown is an atomic flag indicating if the client has been shutdown.
	shutdown int32
	wg       sync.WaitGroup
}

// Connect will attempt to connect to the given Figg node.
func Connect(addr string, options ...Option) (*Figg, error) {
	opts := defaultOptions(addr)
	for _, opt := range options {
		opt(opts)
	}

	figg := &Figg{
		opts:     opts,
		shutdown: 0,
	}
	figg.conn = newConnection(figg.onConnStateChange, opts)
	if err := figg.conn.Connect(); err != nil {
		return nil, err
	}

	figg.wg.Add(1)
	go figg.readLoop()

	figg.wg.Add(1)
	go figg.pingLoop()

	return figg, nil
}

// Publish publishes the data to the given topic. When the server acknowledges
// the message onACK is called.
func (f *Figg) Publish(name string, data []byte, onACK func()) {
	f.conn.Publish(name, data, onACK)
}

// PublishBlocking is similar to Publish except it will block waiting for the
// message is acknowledged. Note this will seriously limit thoughput so if
// high thoughput is needed use Publish and don't wait for messages to be
// acknowledged before sending the next.
func (f *Figg) PublishWaitForACK(name string, data []byte) {
	ch := make(chan interface{}, 1)
	f.conn.Publish(name, data, func() {
		ch <- struct{}{}
	})
	<-ch
}

// PublishNoACK is the same as Publish except it doesn't wait for the message
// to be acknowledged
func (f *Figg) PublishNoACK(name string, data []byte) {
	f.conn.Publish(name, data, nil)
}

// Subscribe to the given topic.
//
// Note only one subscriber is allowed per topic.
func (f *Figg) Subscribe(name string, onMessage MessageCB, options ...TopicOption) error {
	opts := defaultTopicOptions()
	for _, opt := range options {
		opt(opts)
	}

	ch := make(chan interface{}, 1)
	onAttached := func() {
		ch <- struct{}{}
	}
	if opts.FromOffset {
		if err := f.conn.AttachFromOffset(name, opts.Offset, onAttached, onMessage); err != nil {
			return err
		}
	} else {
		if err := f.conn.Attach(name, onAttached, onMessage); err != nil {
			return err
		}
	}
	<-ch
	return nil
}

func (f *Figg) Unsubscribe(topic string) {
	// Note doesn't wait for a response.
	f.conn.Detach(topic)
}

func (f *Figg) Close() error {
	// This will avoid log spam about errors when we shut down.
	atomic.StoreInt32(&f.shutdown, 1)

	if err := f.conn.Close(); err != nil {
		return err
	}

	// Closing the network connection will cause the read loop to exit.
	f.wg.Wait()

	return nil
}

func (f *Figg) readLoop() {
	defer f.wg.Done()

	for {
		if err := f.conn.Recv(); err != nil {
			if s := atomic.LoadInt32(&f.shutdown); s == 1 {
				return
			}

			f.conn.Reconnect()
		}
	}
}

func (f *Figg) pingLoop() {
	defer f.wg.Done()

	for {
		if s := atomic.LoadInt32(&f.shutdown); s == 1 {
			return
		}

		<-time.After(f.opts.PingInterval)
		// Note don't need to reconnect. If the ping expires it will close
		// the connection causing the read loop to reconnect.
		f.conn.Ping()
	}
}

func (f *Figg) onConnStateChange(state ConnState) {
	// Avoid logging if we've been shutdown.
	if s := atomic.LoadInt32(&f.shutdown); s == 1 {
		return
	}

	f.opts.Logger.Debug(
		"connection state change",
		zap.String("state", state.String()),
	)

	if f.opts.ConnStateChangeCB != nil {
		f.opts.ConnStateChangeCB(state)
	}
}
