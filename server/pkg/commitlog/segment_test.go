package commitlog

import (
	"math/rand"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSegment_AppendThenLookup(t *testing.T) {
	segment := NewInMemorySegment(1024, 0)
	assert.Nil(t, segment.Append([]byte("foo")))
	assert.Nil(t, segment.Append([]byte("bar")))
	assert.Nil(t, segment.Append([]byte("car")))

	b, err := segment.Lookup(0)
	assert.Nil(t, err)
	assert.Equal(t, []byte("foo"), b)

	b, err = segment.Lookup(7)
	assert.Nil(t, err)
	assert.Equal(t, []byte("bar"), b)

	b, err = segment.Lookup(14)
	assert.Nil(t, err)
	assert.Equal(t, []byte("car"), b)

	b, err = segment.Lookup(21)
	assert.Equal(t, ErrNotFound, err)
}

func TestSegment_Persist(t *testing.T) {
	dir := "data/" + uuid.New().String()
	defer os.Remove(dir)

	// Use a large segment size so all messages fit in the same segment.
	segment := NewInMemorySegment(1024, 500)
	assert.Nil(t, segment.Append([]byte("foo")))
	assert.Nil(t, segment.Append([]byte("bar")))
	assert.Nil(t, segment.Append([]byte("car")))

	persistedSegment, err := segment.Persist(dir)
	assert.Nil(t, err)

	assert.Equal(t, uint64(500), persistedSegment.Offset())
	assert.Equal(t, uint64(21), persistedSegment.Size())

	b, err := persistedSegment.Lookup(0)
	assert.Nil(t, err)
	assert.Equal(t, []byte("foo"), b)

	b, err = persistedSegment.Lookup(7)
	assert.Nil(t, err)
	assert.Equal(t, []byte("bar"), b)

	b, err = persistedSegment.Lookup(14)
	assert.Nil(t, err)
	assert.Equal(t, []byte("car"), b)

	b, err = persistedSegment.Lookup(21)
	assert.Equal(t, ErrNotFound, err)
}

func benchmarkSegmentPersist(appends int, messageLen int) {
	dir := "data/" + uuid.New().String()
	defer os.RemoveAll(dir)

	message := make([]byte, messageLen)
	rand.Read(message)

	segment := NewInMemorySegment(1<<22, 0)
	for i := 0; i != appends; i++ {
		segment.Append(message)
	}

	if _, err := segment.Persist(dir); err != nil {
		panic(err)
	}
}

func BenchmarkSegment_Persist_Append1000_M1000(b *testing.B) {
	for n := 0; n < b.N; n++ {
		benchmarkSegmentPersist(1000, 1000)
	}
}
