package commitlog

import (
	"math/rand"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCommitLog_AppendThenLookupOneSegment(t *testing.T) {
	dir := "data/" + uuid.New().String()
	defer os.RemoveAll(dir)

	// Use a large segment size so all messages fit in the same segment.
	log := NewCommitLog(100, dir)
	assert.Nil(t, log.Append([]byte("foo")))
	assert.Nil(t, log.Append([]byte("bar")))
	assert.Nil(t, log.Append([]byte("car")))

	b, err := log.Lookup(0)
	assert.Nil(t, err)
	assert.Equal(t, []byte("foo"), b)

	b, err = log.Lookup(7)
	assert.Nil(t, err)
	assert.Equal(t, []byte("bar"), b)

	b, err = log.Lookup(14)
	assert.Nil(t, err)
	assert.Equal(t, []byte("car"), b)

	b, err = log.Lookup(21)
	assert.Equal(t, ErrNotFound, err)
}

func TestCommitLog_AppendThenLookupMultiSegment(t *testing.T) {
	dir := "data/" + uuid.New().String()
	defer os.RemoveAll(dir)

	// Use a small segment size so each message has its own segment.
	log := NewCommitLog(5, dir)
	assert.Nil(t, log.Append([]byte("foo")))
	assert.Nil(t, log.Append([]byte("bar")))
	assert.Nil(t, log.Append([]byte("car")))

	b, err := log.Lookup(0)
	assert.Nil(t, err)
	assert.Equal(t, []byte("foo"), b)

	b, err = log.Lookup(7)
	assert.Nil(t, err)
	assert.Equal(t, []byte("bar"), b)

	b, err = log.Lookup(14)
	assert.Nil(t, err)
	assert.Equal(t, []byte("car"), b)

	b, err = log.Lookup(21)
	assert.Equal(t, ErrNotFound, err)
}

func TestCommitLog_AppendThenLookupPersistedSegment(t *testing.T) {
	dir := "data/" + uuid.New().String()
	defer os.RemoveAll(dir)

	// Use a small segment size so each message has its own segment.
	log := NewCommitLog(5, dir)
	assert.Nil(t, log.Append([]byte("foo")))
	assert.Nil(t, log.Flush())
	assert.Nil(t, log.Append([]byte("bar")))
	assert.Nil(t, log.Flush())
	assert.Nil(t, log.Append([]byte("car")))
	assert.Nil(t, log.Flush())

	b, err := log.Lookup(0)
	assert.Nil(t, err)
	assert.Equal(t, []byte("foo"), b)

	b, err = log.Lookup(7)
	assert.Nil(t, err)
	assert.Equal(t, []byte("bar"), b)

	b, err = log.Lookup(14)
	assert.Nil(t, err)
	assert.Equal(t, []byte("car"), b)

	b, err = log.Lookup(21)
	assert.Equal(t, ErrNotFound, err)
}

func benchmarkCommitLog(appends int, messageLen int) {
	dir := "data/" + uuid.New().String()
	defer os.RemoveAll(dir)

	log := NewCommitLog(1<<24, dir)

	message := make([]byte, messageLen)
	rand.Read(message)

	for i := 0; i != appends; i++ {
		if err := log.Append(message); err != nil {
			panic(err)
		}
	}

	offset := uint64(0)
	for i := 0; i != appends; i++ {
		b, err := log.Lookup(offset)
		if err != nil {
			panic(err)
		}
		if len(b) != messageLen {
			panic("invalid message")
		}
		offset += PrefixSize
		offset += uint64(messageLen)
	}
}

func BenchmarkCommitLog_Append1000_M10(b *testing.B) {
	for n := 0; n < b.N; n++ {
		benchmarkCommitLog(1000, 10)
	}
}

func BenchmarkCommitLog_Append1000_M256KB(b *testing.B) {
	for n := 0; n < b.N; n++ {
		benchmarkCommitLog(1000, 256000)
	}
}
