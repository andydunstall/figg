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

func (t *Topics) OnMessage(topicName string, m []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()

	topic, ok := t.topics[topicName]
	if !ok {
		return
	}

	topic.OnMessage(m)
}

// Subscribes to the given topic. Returns true if this is the first subscriber
// for the topic, false otherwise.
func (t *Topics) Subscribe(topicName string, s MessageSubscriber) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	topic, ok := t.topics[topicName]
	if !ok {
		topic = NewTopic()
		t.topics[topicName] = topic
	}

	return topic.Subscribe(s)
}

func (t *Topics) Unsubscribe(topicName string, s MessageSubscriber) {
	t.mu.Lock()
	defer t.mu.Unlock()

	topic, ok := t.topics[topicName]
	if !ok {
		return
	}

	topic.Unsubscribe(s)
}
