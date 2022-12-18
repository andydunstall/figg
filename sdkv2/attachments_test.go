package figg

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttachments_Pending(t *testing.T) {
	attachments := newAttachments()

	onAttach := func() {}
	attachments.AddPending("foo", onAttach)
	attachments.AddPendingFromOffset("bar", 10, onAttach)
	attachments.AddPending("car", onAttach)

	// After becoming attached the topic 'car' should no longer be pending.
	attachments.OnAttached("car", 20)

	pending := attachments.Pending()

	// Sort as order undefined.
	sort.Slice(pending, func(i, j int) bool {
	  return pending[i].Name < pending[j].Name
	})

	assert.Equal(t, 2, len(pending))

	assert.Equal(t, "bar", pending[0].Name)
	assert.Equal(t, true, pending[0].FromOffset)
	assert.Equal(t, uint64(10), pending[0].Offset)

	assert.Equal(t, "foo", pending[1].Name)
	assert.Equal(t, false, pending[1].FromOffset)
}

// Tests when a pending topic is attached it becomes active.
func TestAttachments_OnAttached(t *testing.T) {
	attachments := newAttachments()

	fooAttached := false
	attachments.AddPending("foo", func() {
		fooAttached = true
	})

	barAttached := false
	attachments.AddPendingFromOffset("bar", 10, func() {
		barAttached = true
	})

	attachments.OnAttached("foo", 20)
	attachments.OnAttached("bar", 10)

	assert.True(t, fooAttached)
	assert.True(t, barAttached)

	assert.Equal(t, 0, len(attachments.Pending()))

	active := attachments.Active()

	// Sort as order undefined.
	sort.Slice(active, func(i, j int) bool {
	  return active[i].Name < active[j].Name
	})

	assert.Equal(t, 2, len(active))

	assert.Equal(t, "bar", active[0].Name)
	assert.Equal(t, uint64(10), active[0].Offset)

	assert.Equal(t, "foo", active[1].Name)
	assert.Equal(t, uint64(20), active[1].Offset)
}
