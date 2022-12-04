package commitlog

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFileSegment_AppendThenLookup(t *testing.T) {
	path := "data/" + uuid.New().String()
	defer os.Remove(path)

	// Use a large segment size so all messages fit in the same segment.
	segment, err := NewFileSegment(path)
	assert.Nil(t, err)
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
