package cluster

import (
	"sync"
)

// Note not currently releasing ports as expect to always have enough.
type PortAllocator struct {
	from uint16
	to   uint16
	next uint16

	mu sync.Mutex
}

func NewPortAllocator(from uint16, to uint16) *PortAllocator {
	return &PortAllocator{
		from: from,
		to:   to,
		next: from,
		mu:   sync.Mutex{},
	}
}

func (a *PortAllocator) Take() uint16 {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.next == a.to {
		// We should always have enough so just panic if we exceed the range.
		panic("port range exceeded")
	}

	port := a.next
	a.next += 1
	return port
}
