package figg

import (
	"io"
)

type reader struct {
	r   io.Reader
	buf []byte
}

func newReader(r io.Reader, bufLen int) *reader {
	return &reader{
		r:   r,
		buf: make([]byte, bufLen),
	}
}

func (r *reader) Read() ([]byte, error) {
	n, err := r.r.Read(r.buf)
	if err != nil {
		return nil, err
	}
	return r.buf[:n], err
}
