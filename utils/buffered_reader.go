package utils

import (
	"io"
)

// BufferedReader reads full protocol messages from the given reader.
//
// This is NOT thread safe.
type BufferedReader struct {
	r io.Reader
	// buf is a buffer to read bytes from the reader into.
	buf []byte
	// pending is a buffer containing partial messages yet to be processed.
	pending []byte
}

func NewBufferedReader(r io.Reader, bufLen int) *BufferedReader {
	return &BufferedReader{
		r:       r,
		buf:     make([]byte, bufLen),
		pending: []byte{},
	}
}

// Read reads a protocol message from the underlying reader, returning the
// message type and payload. This keeps reading more until it has a full
// protocol message.
func (r *BufferedReader) Read() (MessageType, []byte, error) {
	for {
		// If there are pending bytes to process we must process them first.
		if len(r.pending) != 0 {
			messageType, data, ok := r.processBuffer(r.pending, len(r.pending))
			if ok {
				r.pending = r.pending[HeaderLen+len(data):]
				return messageType, data, nil
			}

			// If pending doesn't include a full message keep reading.
		}

		n, err := r.r.Read(r.buf)
		if err != nil {
			return MessageType(0), nil, err
		}

		// If we have pending bytes just append and process them in the next
		// loop, as now we may have a full message.
		if len(r.pending) != 0 {
			r.pending = append(r.pending, r.buf[:n]...)
			continue
		}

		// If we don't already have pending bytes, try to process buf directly,
		// since if it contains a full protocol message this avoids an extra
		// copy to pending. Though if we don't append and process next loop.
		messageType, data, ok := r.processBuffer(r.buf, n)
		if !ok {
			r.pending = append(r.pending, r.buf[:n]...)
			continue
		}
		// If there are still bytes remaining add them to be processed next
		// read.
		r.pending = append(r.pending, r.buf[HeaderLen+len(data):n]...)

		return messageType, data, nil
	}
}

func (r *BufferedReader) processBuffer(buf []byte, bufLen int) (MessageType, []byte, bool) {
	// If the buffer does not contain a full message, it must be a partial so
	// keep reading.
	messageType, payloadLen, ok := DecodeHeader(buf[:bufLen])
	if !ok {
		return MessageType(0), nil, false
	}
	if HeaderLen+payloadLen > bufLen {
		return MessageType(0), nil, false
	}

	return messageType, buf[HeaderLen : HeaderLen+payloadLen], true
}
