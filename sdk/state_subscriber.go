package wombat

// StateSubscriber receives notifications about state changes to the wombat
// client, such as connected, disconnected etc. This will only be called from
// a single goroutine.
type StateSubscriber interface {
	// NotifyState notifies the subscriber about a state change.
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
