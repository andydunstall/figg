package figg

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

func (s *StreamBuffer) Next() ([]byte, bool) {
	if PrefixSize > len(s.buf) {
		return nil, false
	}

	payloadLen := binary.BigEndian.Uint32(s.buf[0:PrefixSize])
	if PrefixSize+int(payloadLen) > len(s.buf) {
		return nil, false
	}

	b := s.buf[PrefixSize : PrefixSize+payloadLen]
	s.buf = s.buf[PrefixSize+payloadLen:]
	return b, true
}
