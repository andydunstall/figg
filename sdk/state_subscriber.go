package figg

import (
	"context"
)

// StateSubscriber receives notifications about state changes to the figg
// client, such as connected, disconnected etc. This will only be called from
// a single goroutine.
type StateSubscriber interface {
	// NotifyState notifies the subscriber about a state change. Note this must
	// not block as is called syncrously.
	NotifyState(state State)
}

// ChannelStateSubscriber subscribes to state updates using a channel. Note
// events must be processed quickly as the event loop will be blocked until
// the event can be sent.
type ChannelStateSubscriber struct {
	ch chan State
}

func NewChannelStateSubscriber() *ChannelStateSubscriber {
	return &ChannelStateSubscriber{
		ch: make(chan State),
	}
}

func (s *ChannelStateSubscriber) Ch() <-chan State {
	return s.ch
}

func (s *ChannelStateSubscriber) NotifyState(state State) {
	s.ch <- state
}

func (s *ChannelStateSubscriber) WaitForConnected(ctx context.Context) error {
	for {
		select {
		case state := <-s.ch:
			if state == StateConnected {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
