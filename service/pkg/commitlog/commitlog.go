package commitlog

type CommitLog struct {
	segment *Segment
}

func NewCommitLog() *CommitLog {
	return &CommitLog{
		segment: NewSegment(),
	}
}

func (s *CommitLog) Offset() uint64 {
	return s.segment.Offset()
}

func (l *CommitLog) Append(b []byte) uint64 {
	return l.segment.Append(b)
}

func (l *CommitLog) Lookup(offset uint64) ([]byte, uint64, bool) {
	return l.segment.Lookup(offset)
}
