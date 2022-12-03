package commitlog

import (
	"io"
	"math/rand"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFileCommitLog_AppendThenLookup(t *testing.T) {
	log, err := NewFileCommitLog("/tmp/" + uuid.New().String())
	assert.Nil(t, err)
	defer log.Remove()

	assert.Nil(t, log.Append([]byte("foo")))
	assert.Nil(t, log.Append([]byte("bar")))
	assert.Nil(t, log.Append([]byte("car")))

	b, offset, err := log.Lookup(0)
	assert.Nil(t, err)
	assert.Equal(t, []byte("foo"), b)

	b, offset, err = log.Lookup(offset)
	assert.Nil(t, err)
	assert.Equal(t, []byte("bar"), b)

	b, offset, err = log.Lookup(offset)
	assert.Nil(t, err)
	assert.Equal(t, []byte("car"), b)

	_, _, err = log.Lookup(offset)
	assert.Equal(t, io.EOF, err)
}

func TestFileCommitLog_LookupEmpty(t *testing.T) {
	log, err := NewFileCommitLog("/tmp/" + uuid.New().String())
	assert.Nil(t, err)
	defer log.Remove()

	_, _, err = log.Lookup(0)
	assert.Equal(t, io.EOF, err)
	_, _, err = log.Lookup(10)
	assert.Equal(t, io.EOF, err)
}

func benchmarkFileCommitLog(appends int, messageLen int) {
	log, err := NewFileCommitLog("/tmp/" + uuid.New().String())
	if err != nil {
		panic(err)
	}
	defer log.Remove()

	message := make([]byte, messageLen)
	rand.Read(message)

	for i := 0; i != appends; i++ {
		if err = log.Append(message); err != nil {
			panic(err)
		}
	}

	var b []byte
	var offset uint64
	for i := 0; i != appends; i++ {
		b, offset, err = log.Lookup(offset)
		if err != nil {
			panic(err)
		}
		if len(b) != messageLen {
			panic("invalid message")
		}
	}
}

func BenchmarkFileCommitLog_Append1000_M10(b *testing.B) {
	for n := 0; n < b.N; n++ {
		benchmarkFileCommitLog(1000, 10)
	}
}

func BenchmarkFileCommitLog_Append1000_M256KB(b *testing.B) {
	for n := 0; n < b.N; n++ {
		benchmarkFileCommitLog(1000, 256000)
	}
}
