package topic

type Subscriptions struct {
	broker        *Broker
	messageCh     chan TopicMessage
	subscriptions map[*Subscription]interface{}
}

func NewSubscriptions(broker *Broker) *Subscriptions {
	return &Subscriptions{
		broker:        broker,
		messageCh:     make(chan TopicMessage),
		subscriptions: make(map[*Subscription]interface{}),
	}
}

func (s *Subscriptions) MessageCh() <-chan TopicMessage {
	return s.messageCh
}

func (s *Subscriptions) AddSubscription(topicName string) {
	topic := s.broker.GetTopic(topicName)
	sub := NewSubscription(s.messageCh, topic)
	s.subscriptions[sub] = struct{}{}
}

func (s *Subscriptions) AddSubscriptionFromOffset(topicName string, lastOffset uint64) {
	topic := s.broker.GetTopic(topicName)
	sub := NewSubscriptionFromOffset(s.messageCh, topic, lastOffset)
	s.subscriptions[sub] = struct{}{}
}

func (s *Subscriptions) Shutdown() {
	for sub, _ := range s.subscriptions {
		sub.Shutdown()
	}
}
