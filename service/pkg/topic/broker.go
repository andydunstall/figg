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

func (b *Broker) GetTopic(name string) (*Topic, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if topic, ok := b.topics[name]; ok {
		return topic, nil
	}
	topic, err := NewTopic(name)
	if err != nil {
		return nil, err
	}
	b.topics[name] = topic
	return topic, nil
}
