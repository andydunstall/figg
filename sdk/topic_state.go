package figg

type TopicState int

const (
	// ATTACHED when the client is connected and is received messages published
	// to the topic.
	ATTACHED = TopicState(iota)
	// ATTACHING when clients connection has dropped and will reattach once
	// connected.
	ATTACHING
	// DETACHED when the user has explicitly unsubscribed from the topic.
	DETACHED
)

func (s TopicState) String() string {
	switch s {
	case ATTACHED:
		return "ATTACHED"
	case ATTACHING:
		return "ATTACHING"
	case DETACHED:
		return "DETACHED"
	default:
		return "UNKNOWN"
	}
}
