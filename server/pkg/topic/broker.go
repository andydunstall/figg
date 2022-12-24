package topic

import (
	"sync"
)

// Broker manages the set of topics active on this node.
type Broker struct {
	// Mutex protecting the below fields.
	mu sync.Mutex

	topics  map[string]*Topic
	options Options
}

func NewBroker(options Options) *Broker {
	return &Broker{
		mu:      sync.Mutex{},
		topics:  map[string]*Topic{},
		options: options,
	}
}

// GetTopic returns the topic with the given name. If the topic is not active it
// is activated and returned.
func (b *Broker) GetTopic(name string) *Topic {
	b.mu.Lock()
	defer b.mu.Unlock()

	if topic, ok := b.topics[name]; ok {
		return topic
	}
	topic := NewTopic(name, b.options)
	b.topics[name] = topic
	return topic
}
