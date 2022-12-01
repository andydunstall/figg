package commitlog

import (
	"encoding/binary"
	"sync"
)

type Segment struct {
	buf []byte
	mu  sync.Mutex
}

func NewSegment() *Segment {
	return &Segment{
		buf: []byte{},
		mu:  sync.Mutex{},
	}
}

func (s *Segment) Offset() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	return uint64(len(s.buf))
}

func (s *Segment) Append(b []byte) uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	metadata := make([]byte, 8)
	binary.BigEndian.PutUint64(metadata, uint64(len(b)))

	s.buf = append(s.buf, metadata...)
	s.buf = append(s.buf, b...)

	return uint64(len(s.buf))
}

func (s *Segment) Lookup(offset uint64) ([]byte, uint64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if offset+8 > uint64(len(s.buf)) {
		return nil, 0, false
	}

	size := binary.BigEndian.Uint64(s.buf[offset : offset+8])
	if offset+8+size > uint64(len(s.buf)) {
		return nil, 0, false
	}

	return s.buf[offset+8 : offset+8+size], offset + 8 + size, true
}
