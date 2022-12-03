package topic

import (
	"fmt"
	"math/rand"
	"testing"
)

type nopAttachment struct {
	DoneCh   chan interface{}
	expected int
	received int
}

func newNopAttachment(expected int) *nopAttachment {
	return &nopAttachment{
		DoneCh:   make(chan interface{}),
		expected: expected,
		received: 0,
	}
}

func (a *nopAttachment) Send(m TopicMessage) {
	a.received++
	if a.expected == a.received {
		close(a.DoneCh)
	}
}

func benchmarkTopicPublish(topicName string, publishes int, subscribers int, messageLen int) {
	broker := NewBroker()

	message := make([]byte, messageLen)
	rand.Read(message)

	attachment := newNopAttachment(publishes)
	subscriptions := NewSubscriptions(broker, attachment)
	for i := 0; i != subscribers; i++ {
		subscriptions.AddSubscription(topicName)
	}

	topic, err := broker.GetTopic(topicName)
	if err != nil {
		panic(err)
	}
	for i := 0; i != publishes; i++ {
		topic.Publish(message)
	}

	<-attachment.DoneCh
}

func benchmarkTopicResume(topicName string, publishes int, messageLen int) {
	broker := NewBroker()

	message := make([]byte, messageLen)
	rand.Read(message)

	attachment := newNopAttachment(publishes)

	topic, err := broker.GetTopic(topicName)
	if err != nil {
		panic(err)
	}
	for i := 0; i != publishes; i++ {
		topic.Publish(message)
	}

	subscriptions := NewSubscriptions(broker, attachment)
	subscriptions.AddSubscriptionFromOffset(topicName, 0)

	<-attachment.DoneCh
}

func BenchmarkTopicPublish_Pub100_Sub1_M10(b *testing.B) {
	for n := 0; n < b.N; n++ {
		topicName := fmt.Sprintf("bench-topic-%d", n)
		benchmarkTopicPublish(topicName, 100, 1, 10)
	}
}

func BenchmarkTopicPublish_Pub100_Sub1000_M10(b *testing.B) {
	for n := 0; n < b.N; n++ {
		topicName := fmt.Sprintf("bench-topic-%d", n)
		benchmarkTopicPublish(topicName, 100, 1000, 10)
	}
}

func BenchmarkTopicPublish_Pub1000_Sub1_M10(b *testing.B) {
	for n := 0; n < b.N; n++ {
		topicName := fmt.Sprintf("bench-topic-%d", n)
		benchmarkTopicPublish(topicName, 1000, 1, 10)
	}
}

func BenchmarkTopicPublish_Pub1000_Sub1000_M10(b *testing.B) {
	for n := 0; n < b.N; n++ {
		topicName := fmt.Sprintf("bench-topic-%d", n)
		benchmarkTopicPublish(topicName, 1000, 1000, 10)
	}
}

func BenchmarkTopicPublish_Pub1000_Sub1_M256KB(b *testing.B) {
	for n := 0; n < b.N; n++ {
		topicName := fmt.Sprintf("bench-topic-%d", n)
		benchmarkTopicPublish(topicName, 1000, 1, 256000)
	}
}

func BenchmarkTopicPublish_Pub1000_Sub1000_M256KB(b *testing.B) {
	for n := 0; n < b.N; n++ {
		topicName := fmt.Sprintf("bench-topic-%d", n)
		benchmarkTopicPublish(topicName, 1000, 1000, 256000)
	}
}

func BenchmarkTopicResume_Pub100_M10(b *testing.B) {
	for n := 0; n < b.N; n++ {
		topicName := fmt.Sprintf("bench-topic-%d", n)
		benchmarkTopicResume(topicName, 100, 10)
	}
}

func BenchmarkTopicResume_Pub1000_M256KB(b *testing.B) {
	for n := 0; n < b.N; n++ {
		topicName := fmt.Sprintf("bench-topic-%d", n)
		benchmarkTopicResume(topicName, 1000, 256000)
	}
}
