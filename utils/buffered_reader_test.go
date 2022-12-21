package utils

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests reading where a network read contains a single message.
func TestBufferedReader_ReadOneToOneMessage(t *testing.T) {
	buf := encodeProtocolMessage(TypeData, []byte("foo"))
	reader := NewBufferedReader(bytes.NewReader(buf), 12)

	messageType, data, err := reader.Read()
	assert.Equal(t, TypeData, messageType)
	assert.Equal(t, []byte("foo"), data)
	assert.Nil(t, err)
}

// Tests reading three messages from the network, where each is fragmented
// such that each network read returns a fraction of the messages.
func TestBufferedReader_ReadFragmentedMessages(t *testing.T) {
	// Use multiple stream read sizes to test different scenarios. Such as
	// reads returning a fraction of each message, reads returning overlapping
	// messages etc.
	for i := 1; i != 40; i++ {
		buf := []byte{}
		buf = append(buf, encodeProtocolMessage(TypeData, []byte("foo"))...)
		buf = append(buf, encodeProtocolMessage(TypeAttach, []byte("bar"))...)
		buf = append(buf, encodeProtocolMessage(TypeAttached, []byte("car"))...)

		reader := NewBufferedReader(bytes.NewReader(buf), i)

		messageType, data, err := reader.Read()
		assert.Equal(t, TypeData, messageType)
		assert.Equal(t, []byte("foo"), data)
		assert.Nil(t, err)

		messageType, data, err = reader.Read()
		assert.Equal(t, TypeAttach, messageType)
		assert.Equal(t, []byte("bar"), data)
		assert.Nil(t, err)

		messageType, data, err = reader.Read()
		assert.Equal(t, TypeAttached, messageType)
		assert.Equal(t, []byte("car"), data)
		assert.Nil(t, err)
	}
}

func encodeProtocolMessage(messageType MessageType, payload []byte) []byte {
	header := make([]byte, HeaderLen)
	EncodeHeader(header, 0, messageType, uint32(len(payload)))

	buf := []byte{}
	buf = append(buf, header...)
	buf = append(buf, payload...)
	return buf
}
