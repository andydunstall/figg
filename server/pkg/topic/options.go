package topic

type Options struct {
	// Persisted indicates the commit log segments should be persisted to disk.
	Persisted bool

	// Dir is the directory to store the commit log segments if persisted.
	Dir string

	// SegmentSize is the size of the commit log segments to use.
	SegmentSize uint64
}
