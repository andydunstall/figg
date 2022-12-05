package commitlog

import (
	"encoding/binary"
	"fmt"
	"os"
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
	s.mu.RLock()
	defer s.mu.RUnlock()

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
	s.mu.RLock()
	defer s.mu.RUnlock()

	return uint64(len(s.buf))
}

func (s *InMemorySegment) Persist(dir string) (Segment, error) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("%s/%d.data", dir, s.Offset())
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	if _, err := file.Write(s.buf); err != nil {
		return nil, err
	}

	// As this point the segment should be immutable so read locking should
	// have no contention.
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Just write the whole segment buffer to disk as for format is the same.
	fileSegment, err := NewFileSegment(file, s.offset, uint64(len(s.buf)))
	if err != nil {
		return nil, err
	}
	return fileSegment, nil
}
