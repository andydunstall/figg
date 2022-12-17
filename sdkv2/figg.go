package figg

import (
	"sync"
	"sync/atomic"

	"go.uber.org/zap"
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
		_, err := f.conn.Read()
		if err != nil {
			if s := atomic.LoadInt32(&f.shutdown); s == 1 {
				return
			}
		}
	}
}

func (f *Figg) onConnStateChange(state ConnState) {
	f.opts.Logger.Debug(
		"connection state change",
		zap.String("state", state.String()),
	)
}
