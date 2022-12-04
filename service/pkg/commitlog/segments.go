package commitlog

import (
	"sync"
)

// Note in the first implementation just keeping a slice of segments. In the
// future can experiment with implementing as a BST to lookup the segment
// faster, though not expecting there to be too many so a slice may still be
// faster.
type Segments struct {
	// Protects the below fields. Using an RWMutex as expecting workload to be
	// read heavy.
	mu       sync.RWMutex
	offsets  []uint64
	segments []Segment
}

func NewSegments() *Segments {
	return &Segments{
		mu:       sync.RWMutex{},
		offsets:  []uint64{},
		segments: []Segment{},
	}
}

// Get returns the segment the given offset falls in.
func (s *Segments) Get(offset uint64) Segment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Must iterate in reverse order or we will always match the first segment.
	for i := len(s.offsets) - 1; i >= 0; i-- {
		off := s.offsets[i]
		if offset >= off {
			return s.segments[i]
		}
	}
	return nil
}

func (s *Segments) Last() Segment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.segments) == 0 {
		return nil
	}
	return s.segments[len(s.segments)-1]
}

func (s *Segments) Add(offset uint64, segment Segment) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.offsets = append(s.offsets, offset)
	s.segments = append(s.segments, segment)
}

func (s *Segments) Swap(newSegment Segment) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, off := range s.offsets {
		if off == newSegment.Offset() {
			if s.segments[i].Offset() != newSegment.Offset() {
				panic("segment offsets don't match")
			}
			s.segments[i] = newSegment
		}
	}
}
