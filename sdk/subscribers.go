package wombat

import (
	"sync"
)

type Subscribers struct {
	subscribers map[string]map[MessageSubscriber]interface{}
	mu          sync.Mutex
}

func NewSubscribers() *Subscribers {
	return &Subscribers{
		subscribers: make(map[string]map[MessageSubscriber]interface{}),
		mu:          sync.Mutex{},
	}
}

func (s *Subscribers) Add(topic string, sub MessageSubscriber) {
	s.mu.Lock()
	defer s.mu.Unlock()

	topicSubscribers, ok := s.subscribers[topic]
	if !ok {
		topicSubscribers = make(map[MessageSubscriber]interface{})
		s.subscribers[topic] = topicSubscribers
	}

	topicSubscribers[sub] = struct{}{}
}

func (s *Subscribers) Topics() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	topics := []string{}
	for topic, _ := range s.subscribers {
		topics = append(topics, topic)
	}
	return topics
}

func (s *Subscribers) OnMessage(topic string, m []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	topicSubscribers, ok := s.subscribers[topic]
	if !ok {
		return
	}

	for sub, _ := range topicSubscribers {
		sub.NotifyMessage(m)
	}
}
