package figg

import (
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap/zapcore"
)

type MessageType uint16

const (
	TypeAttach   = MessageType(1)
	TypeAttached = MessageType(2)
	TypeDetach   = MessageType(3)
	TypeDetached = MessageType(4)
	TypePublish  = MessageType(5)
	TypeACK      = MessageType(6)
	TypePayload  = MessageType(7)
	TypePing     = MessageType(8)
	TypePong     = MessageType(9)

	PrefixSize = 4
)

type AttachMessage struct {
	Topic  string
	Offset string
}

func NewAttachMessage(topic string) *ProtocolMessage {
	return &ProtocolMessage{
		Type: TypeAttach,
		Attach: &AttachMessage{
			Topic: topic,
		},
	}
}

func NewAttachMessageWithOffset(topic string, offset string) *ProtocolMessage {
	return &ProtocolMessage{
		Type: TypeAttach,
		Attach: &AttachMessage{
			Topic:  topic,
			Offset: offset,
		},
	}
}

func (m AttachMessage) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("topic", m.Topic)
	enc.AddString("offset", m.Offset)
	return nil
}

type AttachedMessage struct{}

func (m AttachedMessage) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	return nil
}

type DetachMessage struct {
	Topic string
}

func NewDetachMessage(topic string) *ProtocolMessage {
	return &ProtocolMessage{
		Type: TypeDetach,
		Detach: &DetachMessage{
			Topic: topic,
		},
	}
}

func (m DetachMessage) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("topic", m.Topic)
	return nil
}

type DetachedMessage struct{}

func (m DetachedMessage) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	return nil
}

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

func (m PublishMessage) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("topic", m.Topic)
	enc.AddUint64("seq-num", m.SeqNum)
	enc.AddInt("length", len(m.Payload))
	return nil
}

type ACKMessage struct {
	SeqNum uint64
}

func (m ACKMessage) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddUint64("seq-num", m.SeqNum)
	return nil
}

type PayloadMessage struct {
	Topic   string
	Offset  string
	Message []byte
}

func (m PayloadMessage) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("topic", m.Topic)
	enc.AddString("offset", m.Offset)
	enc.AddInt("length", len(m.Message))
	return nil
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

func (m PingMessage) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt64("timestamp", m.Timestamp)
	return nil
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

func (m PongMessage) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt64("timestamp", m.Timestamp)
	return nil
}

type ProtocolMessage struct {
	Type     MessageType
	Attach   *AttachMessage
	Attached *AttachedMessage
	Detach   *DetachMessage
	Detached *DetachedMessage
	Publish  *PublishMessage
	ACK      *ACKMessage
	Payload  *PayloadMessage
	Ping     *PingMessage
	Pong     *PongMessage
}

func (m *ProtocolMessage) Encode() ([]byte, error) {
	return msgpack.Marshal(m)
}

func (m *ProtocolMessage) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("type", TypeToString(m.Type))

	switch m.Type {
	case TypeAttach:
		enc.AddObject("attach", m.Attach)
	case TypeAttached:
		enc.AddObject("attached", m.Attached)
	case TypePublish:
		enc.AddObject("publish", m.Publish)
	case TypeACK:
		enc.AddObject("ack", m.ACK)
	case TypePayload:
		enc.AddObject("payload", m.Payload)
	case TypePing:
		enc.AddObject("ping", m.Ping)
	case TypePong:
		enc.AddObject("pong", m.Pong)
	}

	return nil
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
