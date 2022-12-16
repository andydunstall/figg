package figg

import (
	"net"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"
)

type Figg struct {
	opts *Options
	conn net.Conn
	// reader reads bytes from the connection.
	reader *reader

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

	conn, err := opts.Dialer.Dial("tcp", addr)
	if err != nil {
		opts.Logger.Error(
			"initial connection failed",
			zap.String("addr", addr),
			zap.Error(err),
		)
		return nil, err
	}
	opts.Logger.Debug("initial connection ok", zap.String("addr", addr))

	figg := &Figg{
		opts:     opts,
		conn:     conn,
		reader:   newReader(conn, opts.ReadBufLen),
		shutdown: 0,
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
		_, err := f.reader.Read()
		if err != nil {
			if s := atomic.LoadInt32(&f.shutdown); s == 1 {
				return
			}

			f.opts.Logger.Warn("connection closed unexpectedly", zap.Error(err))
		}
	}
}
