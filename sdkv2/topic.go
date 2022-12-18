package figg

type Message struct {
	// Data contains the published payload.
	Data []byte

	// Offset is the messages position in the topic. This can be used to recover
	// messages from this offset.
	Offset uint64
}

type MessageCB func(m Message)

type topic struct {
	MessageCB func(m Message)

	// Offset is the offset of the last message recieved on the topic.
	Offset uint64
}
