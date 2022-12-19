package figg

import (
	"encoding/binary"
)

const (
	headerLen = 8

	uint16Len = 2
	uint32Len = 4
	uint64Len = 8

	uint32Max = 0xffffffff

	protocolVersion = uint16(1)

	flagNone      = uint16(0)
	flagUseOffset = uint16(1 << 15)
)

func encodeUint16(buf []byte, offset int, n uint16) int {
	if len(buf) < offset+uint16Len {
		panic("buf too small; cannot encode uint16")
	}

	binary.BigEndian.PutUint16(buf[offset:offset+uint16Len], n)
	return offset + uint16Len
}

func decodeUint16(buf []byte, offset int) (uint16, int) {
	if len(buf) < offset+uint16Len {
		panic("buf too small; cannot encode uint16")
	}

	n := binary.BigEndian.Uint16(buf[offset : offset+uint16Len])
	return n, offset + uint16Len
}

func encodeUint32(buf []byte, offset int, n uint32) int {
	if len(buf) < offset+uint32Len {
		panic("buf too small; cannot encode uint32")
	}

	binary.BigEndian.PutUint32(buf[offset:offset+uint32Len], n)
	return offset + uint32Len
}

func decodeUint32(buf []byte, offset int) (uint32, int) {
	if len(buf) < offset+uint32Len {
		panic("buf too small; cannot encode uint32")
	}

	n := binary.BigEndian.Uint32(buf[offset : offset+uint32Len])
	return n, offset + uint32Len
}

func encodeUint64(buf []byte, offset int, n uint64) int {
	if len(buf) < offset+uint32Len {
		panic("buf too small; cannot encode uint64")
	}

	binary.BigEndian.PutUint64(buf[offset:offset+uint64Len], n)
	return offset + uint64Len
}

func decodeUint64(buf []byte, offset int) (uint64, int) {
	if len(buf) < offset+uint64Len {
		panic("buf too small; cannot encode uint64")
	}

	n := binary.BigEndian.Uint64(buf[offset : offset+uint64Len])
	return n, offset + uint64Len
}

func encodeMessageType(buf []byte, offset int, t MessageType) int {
	return encodeUint16(buf, offset, uint16(t))
}

func decodeMessageType(buf []byte, offset int) (MessageType, int) {
	n, offset := decodeUint16(buf, offset)
	return MessageType(n), offset
}

func encodeBytes(buf []byte, offset int, b []byte) int {
	if len(buf) < offset+len(b) {
		panic("buf too small; cannot encode bytes")
	}

	offset = encodeUint32(buf, offset, uint32(len(b)))
	for i := 0; i != len(b); i++ {
		buf[offset+i] = b[i]
	}
	offset += len(b)
	return offset
}

func encodeHeader(buf []byte, offset int, messageType MessageType, payloadLen uint32) int {
	if len(buf) < headerLen {
		panic("buf too small; cannot encode header")
	}

	offset = encodeUint16(buf, offset, uint16(messageType))
	offset = encodeUint16(buf, offset, protocolVersion)
	offset = encodeUint32(buf, offset, payloadLen)
	return offset
}

func decodeHeader(buf []byte) (MessageType, int, bool) {
	if len(buf) < headerLen {
		return MessageType(0), 0, false
	}

	messageType, offset := decodeMessageType(buf, 0)
	// Protocol version is currently unused.
	_, offset = decodeUint16(buf, offset)
	payloadLen, offset := decodeUint32(buf, offset)

	return messageType, int(payloadLen), true
}

func encodeAttachMessage(topic string) []byte {
	payloadLen := uint16Len + uint32Len + len(topic) + uint64Len

	buf := make([]byte, headerLen+payloadLen)
	offset := encodeHeader(buf, 0, TypeAttach, uint32(payloadLen))

	// Flags.
	flags := flagNone
	encodeUint16(buf, offset, flags)
	// Topic.
	encodeBytes(buf, offset, []byte(topic))
	// Offset (unused as flag not set).
	encodeUint64(buf, offset, 0)

	return buf
}

func encodeAttachFromOffsetMessage(topic string, topicOffset uint64) []byte {
	payloadLen := uint16Len + uint32Len + len(topic) + uint64Len

	buf := make([]byte, headerLen+payloadLen)
	offset := encodeHeader(buf, 0, TypeAttach, uint32(payloadLen))

	// Flags.
	flags := flagUseOffset
	encodeUint16(buf, offset, flags)
	// Topic.
	encodeBytes(buf, offset, []byte(topic))
	// Offset.
	encodeUint64(buf, offset, topicOffset)

	return buf
}

func encodeAttachedMessage(topic string, topicOffset uint64) []byte {
	payloadLen := uint32Len + len(topic) + uint64Len

	buf := make([]byte, headerLen+payloadLen)
	offset := encodeHeader(buf, 0, TypeAttached, uint32(payloadLen))

	// Topic.
	offset = encodeBytes(buf, offset, []byte(topic))
	// Offset.
	encodeUint64(buf, offset, topicOffset)

	return buf
}

func encodeDetachMessage(topic string) []byte {
	payloadLen := uint32Len + len(topic)

	buf := make([]byte, headerLen+payloadLen)
	offset := encodeHeader(buf, 0, TypeDetach, uint32(payloadLen))

	// Topic.
	encodeBytes(buf, offset, []byte(topic))

	return buf
}

func encodeDetachedMessage(topic string) []byte {
	payloadLen := uint32Len + len(topic)

	buf := make([]byte, headerLen+payloadLen)
	offset := encodeHeader(buf, 0, TypeDetached, uint32(payloadLen))

	// Topic.
	encodeBytes(buf, offset, []byte(topic))

	return buf
}

func encodePublishMessage(topic string, seqNum uint64, data []byte) []byte {
	payloadLen := uint32Len + len(topic) + uint64Len + uint32Len + len(data)

	buf := make([]byte, headerLen+payloadLen)
	offset := encodeHeader(buf, 0, TypePublish, uint32(payloadLen))

	encodeBytes(buf, offset, []byte(topic))
	encodeUint64(buf, offset, seqNum)
	encodeBytes(buf, offset, data)

	return buf
}

func encodeACKMessage(seqNum uint64) []byte {
	payloadLen := uint64Len

	buf := make([]byte, headerLen+payloadLen)
	offset := encodeHeader(buf, 0, TypeACK, uint32(payloadLen))

	encodeUint64(buf, offset, seqNum)

	return buf
}
