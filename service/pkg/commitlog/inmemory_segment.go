package commitlog

import (
	"encoding/binary"
	"sync"
)

type InMemorySegment struct {
	offset uint64

	// Protects the below fields.
	mu  sync.RWMutex
	buf []byte
}

func NewInMemorySegment(offset uint64) Segment {
	return &InMemorySegment{
		offset: offset,
		mu:     sync.RWMutex{},
		buf:    []byte{},
	}
}

func (s *InMemorySegment) Append(b []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	prefix := make([]byte, PrefixSize)
	binary.BigEndian.PutUint32(prefix, uint32(len(b)))
	s.buf = append(s.buf, prefix...)
	s.buf = append(s.buf, b...)
	return nil
}

func (s *InMemorySegment) Lookup(offset uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if offset+PrefixSize > uint64(len(s.buf)) {
		return nil, ErrNotFound
	}

	payloadSize := uint64(binary.BigEndian.Uint32(s.buf[offset : offset+PrefixSize]))
	if offset+PrefixSize+payloadSize > uint64(len(s.buf)) {
		return nil, ErrNotFound
	}

	return s.buf[offset+PrefixSize : offset+PrefixSize+payloadSize], nil
}

func (s *InMemorySegment) Offset() uint64 {
	return s.offset
}

func (s *InMemorySegment) Size() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	return uint64(len(s.buf))
}
