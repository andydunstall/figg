package figg

import (
	"sync"
)

type Topics struct {
	topics map[string]*Topic
	mu     sync.Mutex
}

func NewTopics() *Topics {
	return &Topics{
		topics: make(map[string]*Topic),
		mu:     sync.Mutex{},
	}
}

// Topics returns a list of the names of the attached topics.
func (t *Topics) Topics() []string {
	topics := []string{}
	for name, _ := range t.topics {
		topics = append(topics, name)
	}
	return topics
}

func (t *Topics) Offset(topicName string) string {
	t.mu.Lock()
	defer t.mu.Unlock()

	topic, ok := t.topics[topicName]
	if !ok {
		return ""
	}

	return topic.Offset()
}

func (t *Topics) OnMessage(topicName string, m []byte, offset string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	topic, ok := t.topics[topicName]
	if !ok {
		return
	}

	topic.OnMessage(m, offset)
}

// Subscribes to the given topic. Returns true if this is the first subscriber
// for the topic, false otherwise.
func (t *Topics) Subscribe(topicName string, cb MessageHandler) (*MessageSubscriber, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	topic, ok := t.topics[topicName]
	if !ok {
		topic = NewTopic(topicName)
		t.topics[topicName] = topic
	}

	return topic.Subscribe(cb)
}

// Unsubscribes from the given topic. Returns true if the topic now has no
// subscribes (so is inactive), false otherwise.
func (t *Topics) Unsubscribe(topicName string, sub *MessageSubscriber) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	topic, ok := t.topics[topicName]
	if !ok {
		return true
	}

	inactive := topic.Unsubscribe(sub)
	if inactive {
		delete(t.topics, topicName)
	}
	return inactive
}
