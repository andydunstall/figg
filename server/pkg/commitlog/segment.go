package commitlog

const (
	PrefixSize = 4
)

type Segment interface {
	Append(b []byte) error
	Lookup(offset uint64) ([]byte, error)
	// Size returns the number of bytes in the segment.
	Size() uint64
	// Offset returns the starting offset of the segment on the commit log.
	Offset() uint64
	// Persists the segment and returns the persisted segment.
	Persist(dir string) (Segment, error)
}
