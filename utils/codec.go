package utils

import (
	"encoding/binary"
)

const (
	HeaderLen = 8

	uint16Len = 2
	uint32Len = 4
	uint64Len = 8

	uint32Max = 0xffffffff

	protocolVersion = uint16(1)

	FlagNone      = uint16(0)
	FlagUseOffset = uint16(1 << 15)
)

func EncodeUint16(buf []byte, offset int, n uint16) int {
	if len(buf) < offset+uint16Len {
		panic("buf too small; cannot encode uint16")
	}

	binary.BigEndian.PutUint16(buf[offset:offset+uint16Len], n)
	return offset + uint16Len
}

func DecodeUint16(buf []byte, offset int) (uint16, int) {
	if len(buf) < offset+uint16Len {
		panic("buf too small; cannot encode uint16")
	}

	n := binary.BigEndian.Uint16(buf[offset : offset+uint16Len])
	return n, offset + uint16Len
}

func EncodeUint32(buf []byte, offset int, n uint32) int {
	if len(buf) < offset+uint32Len {
		panic("buf too small; cannot encode uint32")
	}

	binary.BigEndian.PutUint32(buf[offset:offset+uint32Len], n)
	return offset + uint32Len
}

func DecodeUint32(buf []byte, offset int) (uint32, int) {
	if len(buf) < offset+uint32Len {
		panic("buf too small; cannot encode uint32")
	}

	n := binary.BigEndian.Uint32(buf[offset : offset+uint32Len])
	return n, offset + uint32Len
}

func EncodeUint64(buf []byte, offset int, n uint64) int {
	if len(buf) < offset+uint32Len {
		panic("buf too small; cannot encode uint64")
	}

	binary.BigEndian.PutUint64(buf[offset:offset+uint64Len], n)
	return offset + uint64Len
}

func DecodeUint64(buf []byte, offset int) (uint64, int) {
	if len(buf) < offset+uint64Len {
		panic("buf too small; cannot encode uint64")
	}

	n := binary.BigEndian.Uint64(buf[offset : offset+uint64Len])
	return n, offset + uint64Len
}

func EncodeMessageType(buf []byte, offset int, t MessageType) int {
	return EncodeUint16(buf, offset, uint16(t))
}

func DecodeMessageType(buf []byte, offset int) (MessageType, int) {
	n, offset := DecodeUint16(buf, offset)
	return MessageType(n), offset
}

func EncodeBytes(buf []byte, offset int, b []byte) int {
	if len(buf) < offset+len(b) {
		panic("buf too small; cannot encode bytes")
	}

	offset = EncodeUint32(buf, offset, uint32(len(b)))
	for i := 0; i != len(b); i++ {
		buf[offset+i] = b[i]
	}
	offset += len(b)
	return offset
}

func EncodeHeader(buf []byte, offset int, messageType MessageType, payloadLen uint32) int {
	if len(buf) < HeaderLen {
		panic("buf too small; cannot encode header")
	}

	offset = EncodeUint16(buf, offset, uint16(messageType))
	offset = EncodeUint16(buf, offset, protocolVersion)
	offset = EncodeUint32(buf, offset, payloadLen)
	return offset
}

func DecodeHeader(buf []byte) (MessageType, int, bool) {
	if len(buf) < HeaderLen {
		return MessageType(0), 0, false
	}

	messageType, offset := DecodeMessageType(buf, 0)
	// Protocol version is currently unused.
	_, offset = DecodeUint16(buf, offset)
	payloadLen, offset := DecodeUint32(buf, offset)

	return messageType, int(payloadLen), true
}

func EncodeAttachMessage(topic string) []byte {
	payloadLen := uint16Len + uint32Len + len(topic) + uint64Len

	buf := make([]byte, HeaderLen+payloadLen)
	offset := EncodeHeader(buf, 0, TypeAttach, uint32(payloadLen))

	// Flags.
	flags := FlagNone
	offset = EncodeUint16(buf, offset, flags)
	// Topic.
	offset = EncodeBytes(buf, offset, []byte(topic))
	// Offset (unused as flag not set).
	offset = EncodeUint64(buf, offset, 0)

	return buf
}

func EncodeAttachFromOffsetMessage(topic string, topicOffset uint64) []byte {
	payloadLen := uint16Len + uint32Len + len(topic) + uint64Len

	buf := make([]byte, HeaderLen+payloadLen)
	offset := EncodeHeader(buf, 0, TypeAttach, uint32(payloadLen))

	// Flags.
	flags := FlagUseOffset
	EncodeUint16(buf, offset, flags)
	// Topic.
	EncodeBytes(buf, offset, []byte(topic))
	// Offset.
	EncodeUint64(buf, offset, topicOffset)

	return buf
}

func EncodeAttachedMessage(topic string, topicOffset uint64) []byte {
	payloadLen := uint32Len + len(topic) + uint64Len

	buf := make([]byte, HeaderLen+payloadLen)
	offset := EncodeHeader(buf, 0, TypeAttached, uint32(payloadLen))

	// Topic.
	offset = EncodeBytes(buf, offset, []byte(topic))
	// Offset.
	EncodeUint64(buf, offset, topicOffset)

	return buf
}

func EncodeDetachMessage(topic string) []byte {
	payloadLen := uint32Len + len(topic)

	buf := make([]byte, HeaderLen+payloadLen)
	offset := EncodeHeader(buf, 0, TypeDetach, uint32(payloadLen))

	// Topic.
	EncodeBytes(buf, offset, []byte(topic))

	return buf
}

func EncodeDetachedMessage(topic string) []byte {
	payloadLen := uint32Len + len(topic)

	buf := make([]byte, HeaderLen+payloadLen)
	offset := EncodeHeader(buf, 0, TypeDetached, uint32(payloadLen))

	// Topic.
	EncodeBytes(buf, offset, []byte(topic))

	return buf
}

func EncodePublishMessage(topic string, seqNum uint64, data []byte) []byte {
	payloadLen := uint32Len + len(topic) + uint64Len + uint32Len + len(data)

	buf := make([]byte, HeaderLen+payloadLen)
	offset := EncodeHeader(buf, 0, TypePublish, uint32(payloadLen))

	offset = EncodeBytes(buf, offset, []byte(topic))
	offset = EncodeUint64(buf, offset, seqNum)
	offset = EncodeBytes(buf, offset, data)

	return buf
}

func EncodeACKMessage(seqNum uint64) []byte {
	payloadLen := uint64Len

	buf := make([]byte, HeaderLen+payloadLen)
	offset := EncodeHeader(buf, 0, TypeACK, uint32(payloadLen))

	EncodeUint64(buf, offset, seqNum)

	return buf
}

func EncodeDataMessage(topic string, topicOffset uint64, data []byte) []byte {
	payloadLen := uint32Len + len(topic) + uint64Len + uint32Len + len(data)

	buf := make([]byte, HeaderLen+payloadLen)
	offset := EncodeHeader(buf, 0, TypeData, uint32(payloadLen))

	offset = EncodeBytes(buf, offset, []byte(topic))
	offset = EncodeUint64(buf, offset, topicOffset)
	EncodeBytes(buf, offset, data)

	return buf
}
