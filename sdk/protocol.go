package wombat

import (
	"github.com/vmihailenco/msgpack/v5"
)

type MessageType uint16

const (
	TypeAttach   = MessageType(1)
	TypeAttached = MessageType(2)
	TypePublish  = MessageType(3)
	TypeAck      = MessageType(4)
	TypeMessage  = MessageType(5)
	TypePing     = MessageType(6)
	TypePong     = MessageType(7)
)

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
	Type MessageType
	Ping *PingMessage
	Pong *PongMessage
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
