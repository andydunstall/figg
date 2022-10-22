package broker

import (
	"sync"

	"github.com/andydunstall/wombat/pkg/topic"
)

type Broker struct {
	topic *topic.Topic
	mu    sync.Mutex
}

func NewBroker() *Broker {
	return &Broker{
		topic: topic.NewTopic(),
		mu:    sync.Mutex{},
	}
}

func (b *Broker) GetTopic() *topic.Topic {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.topic
}
