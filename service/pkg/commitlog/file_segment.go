package commitlog

import (
	"encoding/binary"
	"io"
	"os"
	"sync"
)

type FileSegment struct {
	offset uint64
	file   *os.File

	// Protects the below fields.
	mu   sync.RWMutex
	size uint64
}

func NewFileSegment(file *os.File, offset uint64, size uint64) (Segment, error) {
	return &FileSegment{
		offset: offset,
		file:   file,
		size:   size,
	}, nil
}

func (s *FileSegment) Append(b []byte) error {
	if err := binary.Write(s.file, binary.BigEndian, uint32(len(b))); err != nil {
		return err
	}
	if _, err := s.file.Write(b); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.size += PrefixSize
	s.size += uint64(len(b))
	return nil
}

func (s *FileSegment) Lookup(offset uint64) ([]byte, error) {
	if _, err := s.file.Seek(int64(offset), 0); err != nil {
		if err == io.EOF {
			return nil, ErrNotFound
		}
		return nil, err
	}

	var size uint32
	if err := binary.Read(s.file, binary.BigEndian, &size); err != nil {
		if err == io.EOF {
			return nil, ErrNotFound
		}
		return nil, err
	}

	buf := make([]byte, size)
	_, err := io.ReadFull(s.file, buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (s *FileSegment) Offset() uint64 {
	return s.offset
}

func (s *FileSegment) Size() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.size
}

func (s *FileSegment) Persist(dir string) (Segment, error) {
	// Already persisted so just return self.
	return s, nil
}
