package commitlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSegment_AppendThenLookup(t *testing.T) {
	segment := NewSegment()
	segment.Append([]byte("foo"))
	segment.Append([]byte("bar"))
	segment.Append([]byte("car"))

	b, offset, ok := segment.Lookup(0)
	assert.True(t, ok)
	assert.Equal(t, []byte("foo"), b)

	b, offset, ok = segment.Lookup(offset)
	assert.True(t, ok)
	assert.Equal(t, []byte("bar"), b)

	b, offset, ok = segment.Lookup(offset)
	assert.True(t, ok)
	assert.Equal(t, []byte("car"), b)

	_, _, ok = segment.Lookup(offset)
	assert.False(t, ok)
}

func TestSegment_LookupEmpty(t *testing.T) {
	segment := NewSegment()
	_, _, ok := segment.Lookup(0)
	assert.False(t, ok)
	_, _, ok = segment.Lookup(10)
	assert.False(t, ok)
}
