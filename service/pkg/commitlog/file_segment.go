package commitlog

import (
	"encoding/binary"
	"io"
	"os"
	"sync"
)

type FileSegment struct {
	path   string
	offset uint64
	wrFile *os.File
	rdFile *os.File

	// Protects the below fields.
	mu   sync.RWMutex
	size uint64
}

func NewFileSegment(path string) (Segment, error) {
	wrFile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	rdFile, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &FileSegment{
		path:   path,
		wrFile: wrFile,
		rdFile: rdFile,
		size:   0,
	}, nil
}

func (s *FileSegment) Append(b []byte) error {
	if err := binary.Write(s.wrFile, binary.BigEndian, uint32(len(b))); err != nil {
		return err
	}
	if _, err := s.wrFile.Write(b); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.size += PrefixSize
	s.size += uint64(len(b))
	return nil
}

func (s *FileSegment) Lookup(offset uint64) ([]byte, error) {
	if _, err := s.rdFile.Seek(int64(offset), 0); err != nil {
		if err == io.EOF {
			return nil, ErrNotFound
		}
		return nil, err
	}

	var size uint32
	if err := binary.Read(s.rdFile, binary.BigEndian, &size); err != nil {
		if err == io.EOF {
			return nil, ErrNotFound
		}
		return nil, err
	}

	buf := make([]byte, size)
	_, err := io.ReadFull(s.rdFile, buf)
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
