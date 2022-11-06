package conn

import (
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
)

type MessageType uint16

const (
	TypeAttach   = MessageType(1)
	TypeAttached = MessageType(2)
	TypePublish  = MessageType(3)
	TypeACK      = MessageType(4)
	TypePayload  = MessageType(5)
	TypePing     = MessageType(6)
	TypePong     = MessageType(7)
)

type AttachMessage struct{}

type AttachedMessage struct{}

func NewAttachedMessage() *ProtocolMessage {
	return &ProtocolMessage{
		Type:     TypeAttached,
		Attached: &AttachedMessage{},
	}
}

type PublishMessage struct {
	Topic   string
	Payload []byte
}

type ACKMessage struct{}

type PayloadMessage struct {
	Topic   string
	Offset  uint64
	Message []byte
}

func NewPayloadMessage(topic string, offset uint64, m []byte) *ProtocolMessage {
	return &ProtocolMessage{
		Type: TypePayload,
		Payload: &PayloadMessage{
			Topic:   topic,
			Offset:  offset,
			Message: m,
		},
	}
}

type PingMessage struct {
	// Timestamp is the time in milliseconds the ping message was sent.
	Timestamp int64
}

func NewPingMessage(timestamp int64) *ProtocolMessage {
	return &ProtocolMessage{
		Type: TypePing,
		Ping: &PingMessage{
			Timestamp: timestamp,
		},
	}
}

type PongMessage struct {
	// Timestamp echos back the timestamp from the corresponding ping message.
	Timestamp int64
}

func NewPongMessage(timestamp int64) *ProtocolMessage {
	return &ProtocolMessage{
		Type: TypePong,
		Pong: &PongMessage{
			Timestamp: timestamp,
		},
	}
}

type ProtocolMessage struct {
	Type     MessageType
	Attach   *AttachMessage
	Attached *AttachedMessage
	Publish  *PublishMessage
	ACK      *ACKMessage
	Payload  *PayloadMessage
	Ping     *PingMessage
	Pong     *PongMessage
}

func (m *ProtocolMessage) Encode() ([]byte, error) {
	return msgpack.Marshal(m)
}

func ProtocolMessageFromBytes(b []byte) (*ProtocolMessage, error) {
	var m ProtocolMessage
	if err := msgpack.Unmarshal(b, &m); err != nil {
		return nil, err
	}

	// Check each message has the payload corresponding to the message type.
	switch m.Type {
	case TypeAttach:
		if m.Attach == nil {
			return nil, fmt.Errorf("missing message payload")
		}
	case TypeAttached:
		if m.Attached == nil {
			return nil, fmt.Errorf("missing message payload")
		}
	case TypePublish:
		if m.Publish == nil {
			return nil, fmt.Errorf("missing message payload")
		}
	case TypeACK:
		if m.ACK == nil {
			return nil, fmt.Errorf("missing message payload")
		}
	case TypePayload:
		if m.Payload == nil {
			return nil, fmt.Errorf("missing message payload")
		}
	case TypePing:
		if m.Ping == nil {
			return nil, fmt.Errorf("missing message payload")
		}
	case TypePong:
		if m.Pong == nil {
			return nil, fmt.Errorf("missing message payload")
		}
	default:
		return nil, fmt.Errorf("unknown type")
	}

	return &m, nil
}
