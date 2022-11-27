package topic

import (
	"fmt"
	"testing"
)

type nopAttachment struct {
}

func newNopAttachment() Attachment {
	return &nopAttachment{}
}

func (a *nopAttachment) Send(m TopicMessage) {}

func benchmarkTopic(topicName string, publishes int, subscribers int) {
	broker := NewBroker()

	subscriptions := NewSubscriptions(broker, newNopAttachment())
	for i := 0; i != subscribers; i++ {
		subscriptions.AddSubscription(topicName)
	}

	topic := broker.GetTopic(topicName)
	for i := 0; i != publishes; i++ {
		topic.Publish([]byte(fmt.Sprintf("message-%d", i)))
	}
}

func BenchmarkTopic_Pub100_Sub1(b *testing.B) {
	for n := 0; n < b.N; n++ {
		topicName := fmt.Sprintf("bench-topic-%d", n)
		benchmarkTopic(topicName, 100, 1)
	}
}

func BenchmarkTopic_Pub100_Sub1000(b *testing.B) {
	for n := 0; n < b.N; n++ {
		topicName := fmt.Sprintf("bench-topic-%d", n)
		benchmarkTopic(topicName, 100, 1000)
	}
}

func BenchmarkTopic_Pub1000_Sub1(b *testing.B) {
	for n := 0; n < b.N; n++ {
		topicName := fmt.Sprintf("bench-topic-%d", n)
		benchmarkTopic(topicName, 1000, 1)
	}
}

func BenchmarkTopic_Pub1000_Sub1000(b *testing.B) {
	for n := 0; n < b.N; n++ {
		topicName := fmt.Sprintf("bench-topic-%d", n)
		benchmarkTopic(topicName, 1000, 1000)
	}
}
