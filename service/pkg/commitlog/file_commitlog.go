package commitlog

// Note currently only contains a single segment and never clears messages.
type FileCommitLog struct {
	segment *FileSegment
}

func NewFileCommitLog(path string) (*FileCommitLog, error) {
	segment, err := NewFileSegment(path)
	if err != nil {
		return nil, err
	}
	return &FileCommitLog{
		segment: segment,
	}, nil
}

func (l *FileCommitLog) Append(b []byte) error {
	return l.segment.Append(b)
}

func (l *FileCommitLog) Lookup(offset uint64) ([]byte, uint64, error) {
	return l.segment.Lookup(offset)
}

func (l *FileCommitLog) Remove() error {
	return l.segment.Remove()
}

func (l *FileCommitLog) Close() error {
	return l.segment.Close()
}
