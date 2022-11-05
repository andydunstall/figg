package wombat

import (
	"time"

	"go.uber.org/zap"
)

type Config struct {
	// Addr is the address of the wombat service.
	Addr string

	// StateSubscriber subscribes to events about the current state of the
	// client.
	StateSubscriber StateSubscriber

	// PingInterval is the duration between sending pings to monitor the
	// connection is ok. If the client doesn't receive a PONG before it next
	// sends a ping it reconnects. Defaults to 5 seconds.
	PingInterval time.Duration

	// Logger is a custom logger for the client. If nil no log output is
	// created.
	Logger *zap.Logger
}
