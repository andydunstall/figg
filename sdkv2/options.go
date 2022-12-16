package figg

import (
	"net"
	"time"

	"go.uber.org/zap"
)

type Dialer interface {
	Dial(network string, address string) (net.Conn, error)
}

type Options struct {
	// Addr is the address of the Figg node.
	Addr string

	// ReadBufLen is the size of the read buffer ontop of the socket.
	ReadBufLen int

	// Dialer is a custom dialer to connect to the server. If nil uses
	// net.Dialer with a 5 second timeout.
	Dialer Dialer

	// Logger is a custom logger to log events, which should be configured with
	// the desired logging level. If nil no logging is used.
	Logger *zap.Logger
}

type Option func(*Options)

func WithDialer(dialer Dialer) Option {
	return func(opts *Options) {
		opts.Dialer = dialer
	}
}

func WithReadBufLen(readBufLen int) Option {
	return func(opts *Options) {
		opts.ReadBufLen = readBufLen
	}
}

func WithLogger(logger *zap.Logger) Option {
	return func(opts *Options) {
		opts.Logger = logger
	}
}

func defaultOptions(addr string) *Options {
	return &Options{
		Addr:       addr,
		ReadBufLen: 1 << 15, // 32 KB
		Dialer: &net.Dialer{
			Timeout: time.Second * 5,
		},
		Logger: zap.NewNop(),
	}
}
