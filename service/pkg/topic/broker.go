package topic

import (
	"sync"
)

// Broker manages the set of topics active on this node.
type Broker struct {
	// Mutex protecting the below fields.
	mu sync.Mutex

	topics map[string]*Topic
	dir    string
}

func NewBroker(dir string) *Broker {
	return &Broker{
		mu:     sync.Mutex{},
		topics: map[string]*Topic{},
		dir:    dir,
	}
}

// GetTopic returns the topic with the given name. If the topic is not active it
// is activated and returned.
func (b *Broker) GetTopic(name string) (*Topic, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if topic, ok := b.topics[name]; ok {
		return topic, nil
	}
	topic, err := NewTopic(name, b.dir)
	if err != nil {
		return nil, err
	}
	b.topics[name] = topic
	return topic, nil
}
