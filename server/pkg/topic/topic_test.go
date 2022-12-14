package topic

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/andydunstall/figg/server/pkg/commitlog"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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

func (a *nopAttachment) Send(ctx context.Context, m Message) {
	a.received++
	if a.expected == a.received {
		close(a.DoneCh)
	}
}

func TestTopic_PublishMultipleMessages(t *testing.T) {
	topic := NewTopic("mytopic", Options{
		Persisted:   false,
		SegmentSize: 1000,
	})

	topic.Publish([]byte("foo"))
	topic.Publish([]byte("bar"))
	topic.Publish([]byte("car"))

	b, err := topic.GetMessage(0)
	assert.Nil(t, err)
	assert.Equal(t, []byte("foo"), b)

	b, err = topic.GetMessage(7)
	assert.Nil(t, err)
	assert.Equal(t, []byte("bar"), b)

	b, err = topic.GetMessage(14)
	assert.Nil(t, err)
	assert.Equal(t, []byte("car"), b)

	_, err = topic.GetMessage(21)
	assert.Equal(t, commitlog.ErrNotFound, err)
}

func TestTopic_PublishOneMessage(t *testing.T) {
	topic := NewTopic("mytopic", Options{
		Persisted:   false,
		SegmentSize: 1000,
	})

	topic.Publish([]byte("foo"))

	b, err := topic.GetMessage(0)
	assert.Nil(t, err)
	assert.Equal(t, []byte("foo"), b)
}

func TestTopic_GetInitialMessage(t *testing.T) {
	topic := NewTopic("mytopic", Options{
		Persisted:   false,
		SegmentSize: 1000,
	})

	_, err := topic.GetMessage(topic.Offset())
	assert.Equal(t, commitlog.ErrNotFound, err)
}

func benchmarkTopicPublish(topicName string, publishes int, subscribers int, messageLen int) {
	dir := "data/" + uuid.New().String()
	defer os.RemoveAll(dir)

	broker := NewBroker(Options{
		Persisted:   true,
		SegmentSize: 1 << 22,
		Dir:         dir,
	})

	message := make([]byte, messageLen)
	rand.Read(message)

	attachment := newNopAttachment(publishes)
	subscriptions := NewSubscriptions(broker, attachment)
	for i := 0; i != subscribers; i++ {
		subscriptions.AddSubscription(topicName)
	}

	topic := broker.GetTopic(topicName)
	for i := 0; i != publishes; i++ {
		topic.Publish(message)
	}

	<-attachment.DoneCh
}

func benchmarkTopicResume(topicName string, publishes int, messageLen int) {
	dir := "data/" + uuid.New().String()
	defer os.RemoveAll(dir)

	broker := NewBroker(Options{
		Persisted:   true,
		SegmentSize: 1 << 22,
		Dir:         dir,
	})

	message := make([]byte, messageLen)
	rand.Read(message)

	attachment := newNopAttachment(publishes)

	topic := broker.GetTopic(topicName)
	for i := 0; i != publishes; i++ {
		topic.Publish(message)
	}

	subscriptions := NewSubscriptions(broker, attachment)
	subscriptions.AddSubscriptionFromOffset(topicName, 0)

	<-attachment.DoneCh
}

func BenchmarkTopicPublish_Pub1000_Sub1_M1K(b *testing.B) {
	for n := 0; n < b.N; n++ {
		topicName := fmt.Sprintf("bench-topic-%d", n)
		benchmarkTopicPublish(topicName, 1000, 1, 1<<10)
	}
}

func BenchmarkTopicPublish_Pub1000_Sub1000_M1K(b *testing.B) {
	for n := 0; n < b.N; n++ {
		topicName := fmt.Sprintf("bench-topic-%d", n)
		benchmarkTopicPublish(topicName, 1000, 1000, 1<<10)
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

func BenchmarkTopicResume_Pub100_M256KB(b *testing.B) {
	for n := 0; n < b.N; n++ {
		topicName := fmt.Sprintf("bench-topic-%d", n)
		benchmarkTopicResume(topicName, 100, 256000)
	}
}
