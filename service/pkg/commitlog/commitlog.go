package commitlog

import (
	"errors"
	"fmt"
	"os"
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
		segment = NewInMemorySegment(0)
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
		c.segments.Add(segmentOffset, NewInMemorySegment(segmentOffset))
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
	if err := os.MkdirAll(c.dir, os.ModePerm); err != nil {
		return err
	}

	path := fmt.Sprintf("%s/%d.data", c.dir, s.Offset())
	fileSegment, err := NewFileSegment(path)
	if err != nil {
		return err
	}

	// Note iterating messages incrementally to avoid locking for too long.
	// Really at this point we know the segment is immutable so could do without
	// locks.
	offset := uint64(0)
	for {
		m, err := s.Lookup(offset)
		if err == ErrNotFound {
			// Once we've added all the messages in the segment we're done
			break
		}
		offset += PrefixSize
		offset += uint64(len(m))

		if err = fileSegment.Append(m); err != nil {
			return err
		}
	}

	// Swap the old segment with the new persisted segment.
	c.segments.Swap(fileSegment)
	return nil
}
