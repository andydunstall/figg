package utils

import (
	"encoding/binary"
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

	HeaderLen = 8
)

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

func MessageHeader(messageType MessageType, payloadLen uint32) []byte {
	header := make([]byte, HeaderLen)
	binary.BigEndian.PutUint16(header[:2], uint16(messageType))
	binary.BigEndian.PutUint32(header[4:8], uint32(payloadLen))
	return header
}

func AttachMessage(topic string, offset string) []byte {
	topicPrefix := make([]byte, 2)
	binary.BigEndian.PutUint16(topicPrefix, uint16(len(topic)))

	offsetPrefix := make([]byte, 2)
	binary.BigEndian.PutUint16(offsetPrefix, uint16(len(offset)))

	messageLen := uint32(2 + len(topic) + 2 + len(offset))

	buf := make([]byte, 0, HeaderLen+messageLen)
	buf = append(buf, messageHeader(TypeAttach, messageLen)...)
	buf = append(buf, topicPrefix...)
	buf = append(buf, []byte(topic)...)
	buf = append(buf, offsetPrefix...)
	buf = append(buf, []byte(offset)...)

	return buf
}

func PublishMessage(topic string, seqNum uint64, payload []byte) []byte {
	topicPrefix := make([]byte, 2)
	binary.BigEndian.PutUint16(topicPrefix, uint16(len(topic)))

	seqNumEnc := make([]byte, 8)
	binary.BigEndian.PutUint64(seqNumEnc, seqNum)

	payloadPrefix := make([]byte, 4)
	binary.BigEndian.PutUint32(payloadPrefix, uint32(len(payload)))

	messageLen := uint32(2 + len(topic) + 8 + 4 + len(payload))

	buf := make([]byte, 0, HeaderLen+messageLen)
	buf = append(buf, messageHeader(TypePublish, messageLen)...)
	buf = append(buf, topicPrefix...)
	buf = append(buf, []byte(topic)...)
	buf = append(buf, seqNumEnc...)
	buf = append(buf, payloadPrefix...)
	buf = append(buf, payload...)

	return buf
}

func PingMessage(timestamp uint64) []byte {
	timestampEnc := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampEnc, timestamp)

	buf := make([]byte, 0, HeaderLen+8)
	buf = append(buf, messageHeader(TypePing, 8)...)
	buf = append(buf, timestampEnc...)

	return buf
}

func messageHeader(messageType MessageType, payloadLen uint32) []byte {
	header := make([]byte, HeaderLen)
	binary.BigEndian.PutUint16(header[:2], uint16(messageType))
	binary.BigEndian.PutUint32(header[4:8], uint32(payloadLen))
	return header
}
