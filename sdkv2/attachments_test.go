package figg

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttachments_Attaching(t *testing.T) {
	attachments := newAttachments()

	onAttach := func() {}
	attachments.AddAttaching("foo", onAttach)
	attachments.AddAttachingFromOffset("bar", 10, onAttach)
	attachments.AddAttaching("car", onAttach)

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
	})

	barAttached := false
	attachments.AddAttachingFromOffset("bar", 10, func() {
		barAttached = true
	})

	attachments.OnAttached("foo", 20)
	attachments.OnAttached("bar", 10)

	assert.True(t, fooAttached)
	assert.True(t, barAttached)

	assert.Equal(t, 0, len(attachments.Attaching()))

	Attached := attachments.Attached()

	// Sort as order undefined.
	sort.Slice(Attached, func(i, j int) bool {
		return Attached[i].Name < Attached[j].Name
	})

	assert.Equal(t, 2, len(Attached))

	assert.Equal(t, "bar", Attached[0].Name)
	assert.Equal(t, uint64(10), Attached[0].Offset)

	assert.Equal(t, "foo", Attached[1].Name)
	assert.Equal(t, uint64(20), Attached[1].Offset)
}
