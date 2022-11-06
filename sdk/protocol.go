package figg

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

type AttachMessage struct {
	Topic string
}

func NewAttachMessage(topic string) *ProtocolMessage {
	return &ProtocolMessage{
		Type: TypeAttach,
		Attach: &AttachMessage{
			Topic: topic,
		},
	}
}

type AttachedMessage struct{}

type PublishMessage struct {
	Topic   string
	SeqNum  uint64
	Payload []byte
}

func NewPublishMessage(topic string, seqNum uint64, payload []byte) *ProtocolMessage {
	return &ProtocolMessage{
		Type: TypePublish,
		Publish: &PublishMessage{
			Topic:   topic,
			SeqNum:  seqNum,
			Payload: payload,
		},
	}
}

type ACKMessage struct {
	SeqNum uint64
}

type PayloadMessage struct {
	Topic   string
	Offset  uint64
	Message []byte
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

func TypeToString(t MessageType) string {
	switch t {
	case TypeAttach:
		return "ATTACH"
	case TypeAttached:
		return "ATTACHED"
	case TypePublish:
		return "PUBLISH"
	case TypeACK:
		return "ACK"
	case TypePayload:
		return "PAYLOAD"
	case TypePing:
		return "PING"
	case TypePong:
		return "PONG"
	default:
		return "UNKNOWN"
	}
}
