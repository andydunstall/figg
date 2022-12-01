package topic

import (
	"fmt"
	"math/rand"
	"testing"
)

type nopAttachment struct {
}

func newNopAttachment() Attachment {
	return &nopAttachment{}
}

func (a *nopAttachment) Send(m TopicMessage) {}

func benchmarkTopic(topicName string, publishes int, subscribers int, messageLen int) {
	broker := NewBroker()

	message := make([]byte, messageLen)
	rand.Read(message)

	subscriptions := NewSubscriptions(broker, newNopAttachment())
	for i := 0; i != subscribers; i++ {
		subscriptions.AddSubscription(topicName)
	}

	topic := broker.GetTopic(topicName)
	for i := 0; i != publishes; i++ {
		topic.Publish(message)
	}
}

func BenchmarkTopic_Pub100_Sub1_M10(b *testing.B) {
	for n := 0; n < b.N; n++ {
		topicName := fmt.Sprintf("bench-topic-%d", n)
		benchmarkTopic(topicName, 100, 1, 10)
	}
}

func BenchmarkTopic_Pub100_Sub1000_M10(b *testing.B) {
	for n := 0; n < b.N; n++ {
		topicName := fmt.Sprintf("bench-topic-%d", n)
		benchmarkTopic(topicName, 100, 1000, 10)
	}
}

func BenchmarkTopic_Pub1000_Sub1_M10(b *testing.B) {
	for n := 0; n < b.N; n++ {
		topicName := fmt.Sprintf("bench-topic-%d", n)
		benchmarkTopic(topicName, 1000, 1, 10)
	}
}

func BenchmarkTopic_Pub1000_Sub1000_M10(b *testing.B) {
	for n := 0; n < b.N; n++ {
		topicName := fmt.Sprintf("bench-topic-%d", n)
		benchmarkTopic(topicName, 1000, 1000, 10)
	}
}

func BenchmarkTopic_Pub1000_Sub1_M256KB(b *testing.B) {
	for n := 0; n < b.N; n++ {
		topicName := fmt.Sprintf("bench-topic-%d", n)
		benchmarkTopic(topicName, 1000, 1, 256000)
	}
}

func BenchmarkTopic_Pub1000_Sub1000_M256KB(b *testing.B) {
	for n := 0; n < b.N; n++ {
		topicName := fmt.Sprintf("bench-topic-%d", n)
		benchmarkTopic(topicName, 1000, 1000, 256000)
	}
}
