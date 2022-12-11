package commitlog

import (
	"errors"
)

var (
	ErrNotFound = errors.New("not found")
)

type CommitLog struct {
	segments    *Segments
	segmentSize uint64
	dir         string
}

func NewCommitLog(segmentSize uint64, dir string) *CommitLog {
	return &CommitLog{
		segments:    NewSegments(),
		segmentSize: segmentSize,
		dir:         dir,
	}
}

func (c *CommitLog) Append(b []byte) error {
	segment := c.segments.Last()
	if segment == nil {
		segment = NewInMemorySegment(c.segmentSize, 0)
		c.segments.Add(0, segment)
	}

	segment.Append(b)

	if segment.Size() > c.segmentSize {
		// TODO(AD) look at blocking if the number of segments being persisted
		// exceeds N (otherwise if appends exceed the rate we can
		// persist will end up leaking memory)
		go func() {
			if err := c.persist(segment); err != nil {
				panic(err)
			}
		}()
		segmentOffset := segment.Offset() + segment.Size()
		c.segments.Add(segmentOffset, NewInMemorySegment(c.segmentSize, segmentOffset))
	}

	return nil
}

func (c *CommitLog) Lookup(offset uint64) ([]byte, error) {
	segment := c.segments.Get(offset)
	if segment == nil {
		return nil, ErrNotFound
	}

	segmentOffset := offset - segment.Offset()
	return segment.Lookup(segmentOffset)
}

// persist swaps the given segment with a persisted file segment.
func (c *CommitLog) persist(s Segment) error {
	fileSegment, err := s.Persist(c.dir)
	if err != nil {
		return err
	}
	c.segments.Swap(fileSegment)
	return nil
}
