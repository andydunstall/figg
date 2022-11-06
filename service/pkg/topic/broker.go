package topic

import (
	"sync"
)

type Broker struct {
	topics map[string]*Topic
	mu     sync.Mutex
}

func NewBroker() *Broker {
	return &Broker{
		topics: map[string]*Topic{},
		mu:     sync.Mutex{},
	}
}

func (b *Broker) GetTopic(name string) *Topic {
	b.mu.Lock()
	defer b.mu.Unlock()

	if topic, ok := b.topics[name]; ok {
		return topic
	}
	topic := NewTopic(name)
	b.topics[name] = topic
	return topic
}
