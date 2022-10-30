package wombat

import (
	"github.com/vmihailenco/msgpack/v5"
)

type MessageType uint16

const (
	TypePublishMessage = MessageType(1)
	TypeTopicMessage   = MessageType(2)
)

type PublishMessage struct {
	Message []byte
}

type TopicMessage struct {
	Offset  uint64
	Message []byte
}

type ProtocolMessage struct {
	Type           MessageType
	PublishMessage *PublishMessage
	TopicMessage   *TopicMessage
}

func (m *ProtocolMessage) Encode() ([]byte, error) {
	return msgpack.Marshal(m)
}

func ProtocolMessageFromBytes(b []byte) (*ProtocolMessage, error) {
	var m ProtocolMessage
	if err := msgpack.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
