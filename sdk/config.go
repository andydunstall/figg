package wombat

import (
	"go.uber.org/zap"
)

type Config struct {
	// Addr is the address of the wombat service.
	Addr string

	// StateSubscriber subscribes to events about the current state of the
	// client.
	StateSubscriber StateSubscriber

	// Logger is a custom logger for the client. If nil no log output is
	// created.
	Logger *zap.Logger
}
