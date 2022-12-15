package figg

import (
	"net"
	"time"
)

type Dialer interface {
	Dial(network string, address string) (net.Conn, error)
}

type Options struct {
	// Addr is the address of the Figg node.
	Addr string

	// Dialer is a custom dialer to connect to the server. If nil uses
	// net.Dialer with a 5 second timeout.
	Dialer Dialer
}

type Option func(*Options)

func WithDialer(dialer Dialer) Option {
	return func(opts *Options) {
		opts.Dialer = dialer
	}
}

func defaultOptions(addr string) *Options {
	return &Options{
		Addr: addr,
		Dialer: &net.Dialer{
			Timeout: time.Second * 5,
		},
	}
}
