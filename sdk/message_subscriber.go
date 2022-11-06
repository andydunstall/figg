package figg

import (
	"container/list"
	"sync"
)

// MessageSubscriber receives messages from a subscribed topic. This will only
// be called from a single goroutine.
type MessageSubscriber interface {
	// NotifyMessage notifies the subscriber about message. Note this must
	// not block as is called syncrously.
	NotifyMessage(m []byte)
}

// ChannelMessageSubscriber subscribes to messages using a channel. Note
// events must be processed quickly as the event loop will be blocked until
// the event can be sent.
type ChannelMessageSubscriber struct {
	ch chan []byte
}

func NewChannelMessageSubscriber() *ChannelMessageSubscriber {
	return &ChannelMessageSubscriber{
		ch: make(chan []byte),
	}
}

func (s *ChannelMessageSubscriber) Ch() <-chan []byte {
	return s.ch
}

func (s *ChannelMessageSubscriber) NotifyMessage(m []byte) {
	s.ch <- m
}

// queueMessageSubscriber subscribes to messages by appending the received
// messages an an array.
type QueueMessageSubscriber struct {
	messages *list.List
	mu       sync.Mutex
}

func NewQueueMessageSubscriber() *QueueMessageSubscriber {
	return &QueueMessageSubscriber{
		messages: list.New(),
		mu:       sync.Mutex{},
	}
}

func (s *QueueMessageSubscriber) Next() ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.messages.Len() == 0 {
		return nil, false
	}

	m := s.messages.Front()
	s.messages.Remove(m)
	return m.Value.([]byte), true
}

func (s *QueueMessageSubscriber) NotifyMessage(m []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages.PushBack(m)
}
