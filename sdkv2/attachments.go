package figg

import (
	"sync"
)

type pendingAttachment struct {
	Name string
	FromOffset bool
	Offset uint64
	OnAttached func()
}

type activeAttachment struct {
	Name string
	Offset uint64
}

// attachments maintains both the pending and active attachments.
//
// A pending attachment means the attachment is waiting for an ATTACHED response.
// An active attachments has received an ATTACHED response so is receiving
// all messages published to the topic.
type attachments struct {
	// mu is a mutex protecting the below fields
	mu sync.Mutex

	// pending contains attachments that are waiting for an ATTACHED response.
	pending map[string]pendingAttachment
	// active contains attached attachments.
	active map[string]activeAttachment
}

func newAttachments() *attachments {
	return &attachments{
		mu: sync.Mutex{},
		pending: make(map[string]pendingAttachment),
		active: make(map[string]activeAttachment),
	}
}

// AddPending adds a new pending attachment for the topic with the given name.
// When the topic becomes attached the onAttached callback is called.
func (a *attachments) AddPending(name string, onAttached func()) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Don't allow attaching to an already attached topic.
	if _, ok := a.pending[name]; ok {
		return ErrAlreadySubscribed
	}
	if _, ok := a.active[name]; ok {
		return ErrAlreadySubscribed
	}

	a.pending[name] = pendingAttachment{
		Name: name,
		FromOffset: false,
		Offset: 0,
		OnAttached: onAttached,
	}

	return nil
}

// AddPendingFromOffset is the same as AddPending except it requests an offset
// to attach from.
func (a *attachments) AddPendingFromOffset(name string, offset uint64, onAttached func()) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Don't allow attaching to an already attached topic.
	if _, ok := a.pending[name]; ok {
		return ErrAlreadySubscribed
	}
	if _, ok := a.active[name]; ok {
		return ErrAlreadySubscribed
	}

	a.pending[name] = pendingAttachment{
		Name: name,
		FromOffset: true,
		Offset: offset,
		OnAttached: onAttached,
	}

	return nil
}

// Pending returns a list of all pending attachments.
func (a *attachments) Pending() []pendingAttachment {
	a.mu.Lock()
	defer a.mu.Unlock()

	pending := make([]pendingAttachment, 0, len(a.pending))
	for _, att := range a.pending {
		pending = append(pending, att)
	}
	return pending
}

// Active returns a list of all active attachments.
func (a *attachments) Active() []activeAttachment {
	a.mu.Lock()
	defer a.mu.Unlock()

	active := make([]activeAttachment, 0, len(a.active))
	for _, att := range a.active {
		active = append(active, att)
	}
	return active
}

// OnAttached updates the attachments with an ATTACHED response.
//
// This moves pending attachments to active attachments and calls the registered
// onAttached callback.
func (a *attachments) OnAttached(name string, offset uint64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	pending, ok := a.pending[name]
	// If theres no pending attachment for this topic theres nothing to do.
	if !ok {
		return
	}

	pending.OnAttached()
	delete(a.pending, name)

	a.active[name] = activeAttachment{
		Name: name,
		Offset: offset,
	}
}
