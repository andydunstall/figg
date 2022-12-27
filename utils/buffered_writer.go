package utils

import (
	"io"
	"net"
	"sync"
)

// BufferedWriter handles writing to the writer in a background thread to avoid
// blocking Write.
type BufferedWriter struct {
	// mu is a mutex protecting the below fields.
	mu *sync.Mutex
	w  io.Writer
	// queue contains the messages queued to be sent.
	buf [][]byte
	// cv is a condition variable to wait the write loop when there is pending
	// data to write.
	cv     *sync.Cond
	wg     sync.WaitGroup
	closed bool
}

func NewBufferedWriter(w io.Writer) *BufferedWriter {
	mu := &sync.Mutex{}
	writer := &BufferedWriter{
		mu:     mu,
		w:      w,
		buf:    [][]byte{},
		cv:     sync.NewCond(mu),
		wg:     sync.WaitGroup{},
		closed: false,
	}
	writer.wg.Add(1)
	go writer.writeLoop()
	return writer
}

func (w *BufferedWriter) Write(bufs ...[]byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.buf = append(w.buf, bufs...)
	w.cv.Signal()
	return nil
}

func (w *BufferedWriter) Close() error {
	w.mu.Lock()
	w.closed = true
	// Signal the write loop so it closes.
	w.cv.Signal()
	w.mu.Unlock()
	return nil
}

func (w *BufferedWriter) writeLoop() {
	defer w.wg.Done()

	for {
		w.mu.Lock()
		if w.closed {
			return
		}
		// Since we can miss signals when processing the buffer, must only
		// block if buf is empty.
		if len(w.buf) == 0 {
			w.cv.Wait()
		}
		w.mu.Unlock()

		buf := net.Buffers(w.takeBuf())
		if _, err := buf.WriteTo(w.w); err != nil {
			// If we get a write error, expect the server/client will close the
			// connection so exit.
			return
		}
	}
}

func (w *BufferedWriter) takeBuf() [][]byte {
	w.mu.Lock()
	defer w.mu.Unlock()

	buf := w.buf
	w.buf = [][]byte{}
	return buf
}
