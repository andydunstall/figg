package utils

import (
	"encoding/binary"
)

type StreamBuffer struct {
	buf []byte
}

func NewStreamBuffer() *StreamBuffer {
	return &StreamBuffer{
		buf: []byte{},
	}
}

func (s *StreamBuffer) Push(b []byte) {
	s.buf = append(s.buf, b...)
}

func (s *StreamBuffer) Next() (MessageType, []byte, bool) {
	if HeaderLen > len(s.buf) {
		return MessageType(0), nil, false
	}

	messageType := MessageType(binary.BigEndian.Uint16(s.buf[0:2]))
	payloadLen := binary.BigEndian.Uint32(s.buf[4:8])
	if HeaderLen+int(payloadLen) > len(s.buf) {
		return MessageType(0), nil, false
	}

	payload := s.buf[HeaderLen : HeaderLen+payloadLen]
	s.buf = s.buf[HeaderLen+payloadLen:]
	return messageType, payload, true
}
