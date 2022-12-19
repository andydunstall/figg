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

	return figg, nil
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
