package commitlog

import (
	"io"
	"math/rand"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFileSegment_AppendThenLookup(t *testing.T) {
	segment, err := NewFileSegment("/tmp/" + uuid.New().String())
	assert.Nil(t, err)
	defer segment.Remove()

	assert.Nil(t, segment.Append([]byte("foo")))
	assert.Nil(t, segment.Append([]byte("bar")))
	assert.Nil(t, segment.Append([]byte("car")))

	b, offset, err := segment.Lookup(0)
	assert.Nil(t, err)
	assert.Equal(t, []byte("foo"), b)

	b, offset, err = segment.Lookup(offset)
	assert.Nil(t, err)
	assert.Equal(t, []byte("bar"), b)

	b, offset, err = segment.Lookup(offset)
	assert.Nil(t, err)
	assert.Equal(t, []byte("car"), b)

	_, _, err = segment.Lookup(offset)
	assert.Equal(t, io.EOF, err)
}

func TestFileSegment_LookupEmpty(t *testing.T) {
	segment, err := NewFileSegment("/tmp/" + uuid.New().String())
	assert.Nil(t, err)
	defer segment.Remove()

	_, _, err = segment.Lookup(0)
	assert.Equal(t, io.EOF, err)
	_, _, err = segment.Lookup(10)
	assert.Equal(t, io.EOF, err)
}

func benchmarkFileSegment(appends int, messageLen int) {
	segment, err := NewFileSegment("/tmp/" + uuid.New().String())
	if err != nil {
		panic(err)
	}
	defer segment.Remove()

	message := make([]byte, messageLen)
	rand.Read(message)

	for i := 0; i != appends; i++ {
		if err = segment.Append(message); err != nil {
			panic(err)
		}
	}

	var b []byte
	var offset uint64
	for i := 0; i != appends; i++ {
		b, offset, err = segment.Lookup(offset)
		if err != nil {
			panic(err)
		}
		if len(b) != messageLen {
			panic("invalid message")
		}
	}
}

func BenchmarkFileSegment_Append1000_M10(b *testing.B) {
	for n := 0; n < b.N; n++ {
		benchmarkFileSegment(1000, 10)
	}
}

func BenchmarkFileSegment_Append1000_M256KB(b *testing.B) {
	for n := 0; n < b.N; n++ {
		benchmarkFileSegment(1000, 256000)
	}
}
