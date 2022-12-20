package figg

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttachments_Attaching(t *testing.T) {
	attachments := newAttachments()

	onAttach := func() {}
	attachments.AddAttaching("foo", onAttach, nil)
	attachments.AddAttachingFromOffset("bar", 10, onAttach, nil)
	attachments.AddAttaching("car", onAttach, nil)

	// After becoming attached the topic 'car' should no longer be Attaching.
	attachments.OnAttached("car", 20)

	Attaching := attachments.Attaching()

	// Sort as order undefined.
	sort.Slice(Attaching, func(i, j int) bool {
		return Attaching[i].Name < Attaching[j].Name
	})

	assert.Equal(t, 2, len(Attaching))

	assert.Equal(t, "bar", Attaching[0].Name)
	assert.Equal(t, true, Attaching[0].FromOffset)
	assert.Equal(t, uint64(10), Attaching[0].Offset)

	assert.Equal(t, "foo", Attaching[1].Name)
	assert.Equal(t, false, Attaching[1].FromOffset)
}

// Tests when a Attaching topic is attached it becomes Attached.
func TestAttachments_OnAttached(t *testing.T) {
	attachments := newAttachments()

	fooAttached := false
	attachments.AddAttaching("foo", func() {
		fooAttached = true
	}, nil)

	barAttached := false
	attachments.AddAttachingFromOffset("bar", 10, func() {
		barAttached = true
	}, nil)

	attachments.OnAttached("foo", 20)
	attachments.OnAttached("bar", 10)

	assert.True(t, fooAttached)
	assert.True(t, barAttached)

	assert.Equal(t, 0, len(attachments.Attaching()))

	attached := attachments.Attached()

	// Sort as order undefined.
	sort.Slice(attached, func(i, j int) bool {
		return attached[i].Name < attached[j].Name
	})

	assert.Equal(t, 2, len(attached))

	assert.Equal(t, "bar", attached[0].Name)
	assert.Equal(t, uint64(10), attached[0].Offset)

	assert.Equal(t, "foo", attached[1].Name)
	assert.Equal(t, uint64(20), attached[1].Offset)
}

func TestAttachments_AttachingAndAttachedClearedWhenDetaching(t *testing.T) {
	attachments := newAttachments()

	// Add attaching topic.
	attachments.AddAttaching("foo", func() {}, nil)

	// Add attached topic.
	attachments.AddAttaching("bar", func() {}, nil)
	attachments.OnAttached("bar", 10)

	// Make both the above topics detaching. This should remove from attaching
	// and attached.
	attachments.AddDetaching("foo")
	attachments.AddDetaching("bar")

	assert.Equal(t, 0, len(attachments.Attaching()))
	assert.Equal(t, 0, len(attachments.Attached()))
}

func TestAttachments_AttachToDetachingChannel(t *testing.T) {
	attachments := newAttachments()

	// Add attaching topics.
	attachments.AddAttaching("foo", func() {}, nil)
	attachments.AddAttachingFromOffset("bar", 10, func() {}, nil)

	// Replace with detaching topic.
	attachments.AddDetaching("foo")
	attachments.AddDetaching("bar")

	// Attach again before becoming detached.
	attachments.AddAttaching("foo", func() {}, nil)
	attachments.AddAttachingFromOffset("bar", 10, func() {}, nil)

	// Check its not attaching not detaching
	attaching := attachments.Attaching()
	// Sort as order undefined.
	sort.Slice(attaching, func(i, j int) bool {
		return attaching[i].Name < attaching[j].Name
	})
	assert.Equal(t, "bar", attaching[0].Name)
	assert.Equal(t, "foo", attaching[1].Name)

	assert.Equal(t, 0, len(attachments.Detaching()))
}

func TestAttachments_Detaching(t *testing.T) {
	attachments := newAttachments()

	// Add attaching topics.
	attachments.AddAttaching("foo", func() {}, nil)
	attachments.AddAttachingFromOffset("bar", 10, func() {}, nil)

	// Replace with detaching topic.
	attachments.AddDetaching("foo")
	attachments.AddDetaching("bar")

	detaching := attachments.Detaching()
	// Sort as order undefined.
	sort.Strings(detaching)
	assert.Equal(t, []string{"bar", "foo"}, detaching)
}

func TestAttachments_OnDetached(t *testing.T) {
	attachments := newAttachments()

	// Add attaching topics.
	attachments.AddAttaching("foo", func() {}, nil)
	attachments.AddAttachingFromOffset("bar", 10, func() {}, nil)

	// Replace with detaching topic.
	attachments.AddDetaching("foo")
	attachments.AddDetaching("bar")

	// Detach one of the topics.
	attachments.OnDetached("foo")

	detaching := attachments.Detaching()
	assert.Equal(t, []string{"bar"}, detaching)
}

func TestAttachments_OnMessage(t *testing.T) {
	messages := []Message{}
	attachments := newAttachments()

	// Add attached topic.
	attachments.AddAttaching("foo", func() {}, func(m Message) {
		messages = append(messages, m)
	})
	attachments.OnAttached("foo", 10)

	attachments.OnMessage("foo", Message{
		Data:   []byte("A"),
		Offset: 5,
	})
	attachments.OnMessage("foo", Message{
		Data:   []byte("B"),
		Offset: 10,
	})
	attachments.OnMessage("foo", Message{
		Data:   []byte("C"),
		Offset: 15,
	})

	assert.Equal(t, []Message{
		{
			Data:   []byte("A"),
			Offset: 5,
		},
		{
			Data:   []byte("B"),
			Offset: 10,
		},
		{
			Data:   []byte("C"),
			Offset: 15,
		},
	}, messages)

	// Check the tracked offset is updated.
	assert.Equal(t, uint64(15), attachments.Attached()[0].Offset)
}
