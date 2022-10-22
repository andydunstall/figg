package broker

import (
	"sync"

	"github.com/andydunstall/wombat/pkg/topic"
)

type Broker struct {
	topics map[string]*topic.Topic
	mu     sync.Mutex
}

func NewBroker() *Broker {
	return &Broker{
		topics: map[string]*topic.Topic{},
		mu:     sync.Mutex{},
	}
}

func (b *Broker) GetTopic(name string) *topic.Topic {
	b.mu.Lock()
	defer b.mu.Unlock()

	if topic, ok := b.topics[name]; ok {
		return topic
	}
	topic := topic.NewTopic()
	b.topics[name] = topic
	return topic
}
