package topic

type Subscriptions struct {
	broker        *Broker
	attachment    Attachment
	subscriptions map[*Subscription]interface{}
}

func NewSubscriptions(broker *Broker, attachment Attachment) *Subscriptions {
	return &Subscriptions{
		broker:        broker,
		attachment:    attachment,
		subscriptions: make(map[*Subscription]interface{}),
	}
}

func (s *Subscriptions) AddSubscription(topicName string) {
	topic := s.broker.GetTopic(topicName)
	sub := NewSubscription(s.attachment, topic)
	s.subscriptions[sub] = struct{}{}
}

func (s *Subscriptions) AddSubscriptionFromOffset(topicName string, lastOffset uint64) {
	topic := s.broker.GetTopic(topicName)
	sub := NewSubscriptionFromOffset(s.attachment, topic, lastOffset)
	s.subscriptions[sub] = struct{}{}
}

func (s *Subscriptions) Shutdown() {
	for sub, _ := range s.subscriptions {
		sub.Shutdown()
	}
}
