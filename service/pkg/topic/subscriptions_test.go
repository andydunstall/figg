package topic

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubscriptions_SubscribeToMultipleTopics(t *testing.T) {
	broker := NewBroker()
	subs := NewSubscriptions(broker)
	defer subs.Shutdown()

	topics := []string{}
	for i := 0; i != 10; i++ {
		topics = append(topics, fmt.Sprintf("topic-%d", i))
	}

	for _, topicName := range topics {
		subs.AddSubscription(topicName)
	}

	for _, topicName := range topics {
		topic := broker.GetTopic(topicName)
		topic.Publish([]byte("foo"))
	}

	// Verify we received the message on each topic (node order is undefined).
	receivedTopics := []string{}
	for i := 0; i != len(topics); i++ {
		m := <-subs.MessageCh()
		receivedTopics = append(receivedTopics, m.Topic)
	}
	sort.Strings(receivedTopics)
	assert.Equal(t, topics, receivedTopics)
}

func TestSubscriptions_SubscribeToTopic(t *testing.T) {
	broker := NewBroker()
	subs := NewSubscriptions(broker)
	defer subs.Shutdown()

	subs.AddSubscription("foo")

	topic := broker.GetTopic("foo")
	topic.Publish([]byte("bar"))

	assert.Equal(t, TopicMessage{
		Topic:   "foo",
		Offset:  1,
		Message: []byte("bar"),
	}, <-subs.MessageCh())
}
