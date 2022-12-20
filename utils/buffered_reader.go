package utils

import (
	"io"
)

type BufferedReader struct {
	r   io.Reader
	buf []byte
}

func NewBufferedReader(r io.Reader, bufLen int) *BufferedReader {
	return &BufferedReader{
		r:   r,
		buf: make([]byte, bufLen),
	}
}

// Reads bytes from the reader. Must only call from one goroutine.
func (r *BufferedReader) Read() ([]byte, error) {
	n, err := r.r.Read(r.buf)
	if err != nil {
		return nil, err
	}
	return r.buf[:n], err
}
