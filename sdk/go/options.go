package figg

import (
	"math/rand"
	"net"
	"time"

	"go.uber.org/zap"
)

const (
	DefaultReadBufLen   = 1 << 15 // 32 KB
	DefaultWindowSize   = 256
	DefaultPingInterval = 2 * time.Second
	DefaultMaxPingOut   = 2
)

type Dialer interface {
	Dial(network string, address string) (net.Conn, error)
}

type ReconnectBackoffCB func(attempts int) time.Duration

type ConnStateChangeCB func(state ConnState)

type Options struct {
	// Addr is the address of the Figg node.
	Addr string

	// ReadBufLen is the size of the read buffer ontop of the socket.
	ReadBufLen int

	// Dialer is a custom dialer to connect to the server. If nil uses
	// net.Dialer with a 5 second timeout.
	Dialer Dialer

	// ReconnectBackoffCB is a callback to define a custom backoff strategy
	// when attempting to reconnect to the server. If nil uses a default
	// strategy where the retry doubles after each attempt, starting with a
	// 1 second interval after the first attempt, a maximum wait of 30
	// seconds, and adding 20% random jitter (see defaultReconnectBackoffCB).
	ReconnectBackoffCB ReconnectBackoffCB

	// ConnStateChangeCB is an optional callback called when the clients
	// connection state changes. Note this must not block.
	ConnStateChangeCB ConnStateChangeCB

	// WindowSize is the number of unacknowledged in-flight messages are allowed
	// before Publish blocking. Defaults to 256.
	WindowSize int

	// PingInterval is the time between sending pings. Defaults to 2 seconds.
	PingInterval time.Duration

	// MaxPingOut is the maximum number of pings that have not received a
	// pong before determining the connection has dropped. Defaults to 2.
	MaxPingOut int

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

func WithReconnectBackoffCB(cb ReconnectBackoffCB) Option {
	return func(opts *Options) {
		opts.ReconnectBackoffCB = cb
	}
}

func WithConnStateChangeCB(cb ConnStateChangeCB) Option {
	return func(opts *Options) {
		opts.ConnStateChangeCB = cb
	}
}

func WithWindowSize(windowSize int) Option {
	return func(opts *Options) {
		opts.WindowSize = windowSize
	}
}

func WithPingInterval(pingInterval time.Duration) Option {
	return func(opts *Options) {
		opts.PingInterval = pingInterval
	}
}

func WithMaxPingOut(maxPingOut int) Option {
	return func(opts *Options) {
		opts.MaxPingOut = maxPingOut
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
		ReadBufLen: DefaultReadBufLen,
		Dialer: &net.Dialer{
			Timeout: time.Second * 5,
		},
		ReconnectBackoffCB: defaultReconnectBackoffCB,
		ConnStateChangeCB:  nil,
		WindowSize:         DefaultWindowSize,
		PingInterval:       DefaultPingInterval,
		MaxPingOut:         DefaultMaxPingOut,
		Logger:             zap.NewNop(),
	}
}

func defaultReconnectBackoffCB(attempts int) time.Duration {
	// The first time the connection drops retry immediately.
	if attempts == 0 {
		return time.Duration(0)
	}

	delaySeconds := 1 << (attempts - 1)
	if delaySeconds > 30 {
		delaySeconds = 30
	}

	// Add jitter and convert to milliseconds. Such as a delay of 1 second will
	// have a multipler between 800 and 1200 milliseconds.
	minMultiplier := 800
	maxMultiplier := 1200
	jitterMultiplier := rand.Intn(maxMultiplier-minMultiplier) + minMultiplier
	return time.Duration(time.Millisecond * time.Duration(delaySeconds*jitterMultiplier))
}
