package commitlog

import (
	"encoding/binary"
	"io"
	"os"
)

type FileSegment struct {
	path   string
	wrFile *os.File
	rdFile *os.File
}

func NewFileSegment(path string) (*FileSegment, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}

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
	}, nil
}

func (s *FileSegment) Append(b []byte) error {
	if err := binary.Write(s.wrFile, binary.BigEndian, uint32(len(b))); err != nil {
		return err
	}
	if _, err := s.wrFile.Write(b); err != nil {
		return err
	}
	return nil
}

func (s *FileSegment) Lookup(offset uint64) ([]byte, uint64, error) {
	if _, err := s.rdFile.Seek(int64(offset), 0); err != nil {
		return nil, 0, err
	}

	var size uint32
	if err := binary.Read(s.rdFile, binary.BigEndian, &size); err != nil {
		return nil, 0, err
	}

	buf := make([]byte, size)
	_, err := io.ReadFull(s.rdFile, buf)
	if err != nil {
		return nil, 0, err
	}

	return buf, offset + 4 + uint64(size), nil
}

func (s *FileSegment) Remove() error {
	if err := s.Close(); err != nil {
		return err
	}
	if err := os.Remove(s.path); err != nil {
		return err
	}
	return nil
}

func (s *FileSegment) Close() error {
	if err := s.rdFile.Close(); err != nil {
		return err
	}
	if err := s.wrFile.Close(); err != nil {
		return err
	}
	return nil
}
