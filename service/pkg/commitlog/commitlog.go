package commitlog

import (
	"errors"
)

var (
	ErrNotFound = errors.New("not found")
)

// CommitLog maintains the log of messages.
//
// The full design of the commit logs is described in the docs.
type CommitLog struct {
	segments *Segments
	// segmentSize contains the configured size of each segment in the log.
	segmentSize uint64
	// dir is the directory to persist messages to.
	dir string
}

// NewCommitLog creates an empty commit log in the given directory.
func NewCommitLog(segmentSize uint64, dir string) *CommitLog {
	return &CommitLog{
		segments:    NewSegments(),
		segmentSize: segmentSize,
		dir:         dir,
	}
}

// Append adds new data to the commit log. This will be appended to the most
// recent segment, which is in memory so this should be fast.
func (c *CommitLog) Append(b []byte) error {
	segment := c.segments.Last()
	if segment == nil {
		segment = NewInMemorySegment(c.segmentSize, 0)
		c.segments.Add(0, segment)
	}

	segment.Append(b)

	if segment.Size() > c.segmentSize {
		go func() {
			if err := c.persist(segment); err != nil {
				panic(err)
			}
		}()
		c.newSegment()
	}

	return nil
}

// Lookup returns the entry at the given offset in the commit log. If not found
// returns ErrNotFound.
func (c *CommitLog) Lookup(offset uint64) ([]byte, error) {
	segment := c.segments.Get(offset)
	if segment == nil {
		return nil, ErrNotFound
	}

	segmentOffset := offset - segment.Offset()
	return segment.Lookup(segmentOffset)
}

// Flush persists the latest segment to disk (syncrously) and creates a new in-
// memory segment.
//
// Note this will typically occur in the background when the latest segment
// is full, though adding an explicit method for testing.
func (c *CommitLog) Flush() error {
	segment := c.segments.Last()
	if segment == nil {
		// No segments exist so nothing to do.
		return nil
	}

	if err := c.persist(segment); err != nil {
		return err
	}
	c.newSegment()
	return nil
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

// newSegment adds a new empty in-memory segment to the log.
func (c *CommitLog) newSegment() {
	segment := c.segments.Last()
	if segment == nil {
		// First segment so offset is 0.
		c.segments.Add(0, segment)
		return
	}

	segmentOffset := segment.Offset() + segment.Size()
	c.segments.Add(segmentOffset, NewInMemorySegment(c.segmentSize, segmentOffset))
}
