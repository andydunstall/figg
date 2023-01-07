package figg

type Message struct {
	// Data contains the published payload.
	Data []byte

	// Offset is the messages position in the topic. This can be used to recover
	// messages from this offset.
	Offset uint64
}

type MessageCB func(m *Message)
