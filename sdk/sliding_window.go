package figg

import (
	"sync"
)

type unackedMessage struct {
	Topic  string
	Data   []byte
	SeqNum uint64
	OnACK  func()
}

// slidingWindow stores the unacknowledged messages in a circular buffer. When
// full adding new messages will block as a way of controlling Publish.
type slidingWindow struct {
	cv *sync.Cond

	buf  []unackedMessage
	size int
	// head points to the start of the window.
	head int
	// tail points to one past the last element in the window.
	tail int

	seqNum uint64
}

// newSlidingWindow is the maximum number of unacknowledged messages can be
// in-flight before blocking.
func newSlidingWindow(maxSize int) *slidingWindow {
	return &slidingWindow{
		cv:     sync.NewCond(&sync.Mutex{}),
		buf:    make([]unackedMessage, maxSize),
		size:   0,
		head:   0,
		tail:   0,
		seqNum: 0,
	}
}

// Push adds a new message to the window and returns the assigned sequence
// number. If the window is full this will block.
func (w *slidingWindow) Push(topic string, data []byte, onACK func()) uint64 {
	w.cv.L.Lock()
	defer w.cv.L.Unlock()

	seqNum := w.seqNum
	w.seqNum++

	m := unackedMessage{
		Topic:  topic,
		Data:   data,
		SeqNum: seqNum,
		OnACK:  onACK,
	}

	// Block until the window is no longer empty.
	for w.size == len(w.buf) {
		w.cv.Wait()
	}

	// We now know there is room for another element so add.
	w.buf[w.tail] = m
	// Update tail to point to the new item.
	w.tail = (w.tail + 1) % len(w.buf)
	w.size++

	return seqNum
}

// Messages returns all messages in the window in order. This is used to resend
// any unacknowledged messages on reconnect. Note must not modify the returned
// messages.
func (w *slidingWindow) Messages() []unackedMessage {
	w.cv.L.Lock()
	defer w.cv.L.Unlock()

	idx := w.head
	count := 0

	messages := make([]unackedMessage, 0, w.size)
	for count < w.size {
		messages = append(messages, w.buf[idx])
		count++
		idx = (idx + 1) % len(w.buf)
	}

	return messages
}

// Acknowledge acknowledges all messages with a sequence number less than or
// equal to the given sequence number.
func (w *slidingWindow) Acknowledge(seqNum uint64) {
	w.cv.L.Lock()
	defer w.cv.L.Unlock()

	for w.size > 0 && w.buf[w.head].SeqNum <= seqNum {
		if w.buf[w.head].OnACK != nil {
			w.buf[w.head].OnACK()
		}

		w.head = (w.head + 1) % len(w.buf)
		w.size--
	}

	w.cv.Signal()
}
