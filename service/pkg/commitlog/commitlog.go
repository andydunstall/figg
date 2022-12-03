package commitlog

type CommitLog interface {
	Append(b []byte) error
	Lookup(offset uint64) ([]byte, uint64, error)
	Remove() error
	Close() error
}
