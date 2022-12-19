package figg

import (
	"sync"
)

type attachingAttachment struct {
	Name       string
	FromOffset bool
	Offset     uint64
	OnAttached func()
}

type attachedAttachment struct {
	Name   string
	Offset uint64
}

type attachments struct {
	// mu is a mutex protecting the below fields
	mu sync.Mutex

	attaching map[string]attachingAttachment
	attached  map[string]attachedAttachment
	detaching map[string]interface{}
}

func newAttachments() *attachments {
	return &attachments{
		mu:        sync.Mutex{},
		attaching: make(map[string]attachingAttachment),
		attached:  make(map[string]attachedAttachment),
		detaching: make(map[string]interface{}),
	}
}

// AddAttaching adds a new attaching attachment for the topic with the given name.
// When the topic becomes attached the onAttached callback is called.
func (a *attachments) AddAttaching(name string, onAttached func()) error {
	// Don't allow attaching multiple times.
	if a.isAttaching(name) || a.isAttached(name) {
		return ErrAlreadySubscribed
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// If we are trying to detach the topic stop. Otherwise may attach then
	// immediately detach.
	delete(a.detaching, name)

	a.attaching[name] = attachingAttachment{
		Name:       name,
		FromOffset: false,
		Offset:     0,
		OnAttached: onAttached,
	}

	return nil
}

// AddAttachingFromOffset is the same as AddAttaching except it requests an offset
// to attach from.
func (a *attachments) AddAttachingFromOffset(name string, offset uint64, onAttached func()) error {
	// Don't allow attaching multiple times.
	if a.isAttaching(name) || a.isAttached(name) {
		return ErrAlreadySubscribed
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// If we are trying to detach the topic stop. Otherwise may attach then
	// immediately detach.
	delete(a.detaching, name)

	a.attaching[name] = attachingAttachment{
		Name:       name,
		FromOffset: true,
		Offset:     offset,
		OnAttached: onAttached,
	}

	return nil
}

func (a *attachments) AddDetaching(name string) bool {
	// If we're not attaching or attached do nothing.
	if !a.isAttaching(name) && !a.isAttached(name) {
		return false
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.detaching[name] = struct{}{}
	delete(a.attaching, name)
	delete(a.attached, name)
	return true
}

// Attaching returns a list of all attaching attachments.
func (a *attachments) Attaching() []attachingAttachment {
	a.mu.Lock()
	defer a.mu.Unlock()

	attaching := make([]attachingAttachment, 0, len(a.attaching))
	for _, att := range a.attaching {
		attaching = append(attaching, att)
	}
	return attaching
}

// Attached returns a list of all attached attachments.
func (a *attachments) Attached() []attachedAttachment {
	a.mu.Lock()
	defer a.mu.Unlock()

	attached := make([]attachedAttachment, 0, len(a.attached))
	for _, att := range a.attached {
		attached = append(attached, att)
	}
	return attached
}

// Detaching returns a list of all detaching topics.
func (a *attachments) Detaching() []string {
	a.mu.Lock()
	defer a.mu.Unlock()

	detaching := make([]string, 0, len(a.detaching))
	for topic, _ := range a.detaching {
		detaching = append(detaching, topic)
	}
	return detaching
}

// OnAttached updates the attachments with an ATTACHED response.
//
// This moves attaching attachments to attached attachments and calls the registered
// onAttached callback.
func (a *attachments) OnAttached(name string, offset uint64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	attaching, ok := a.attaching[name]
	// If theres no attaching attachment for this topic theres nothing to do.
	if !ok {
		return
	}

	attaching.OnAttached()
	delete(a.attaching, name)

	a.attached[name] = attachedAttachment{
		Name:   name,
		Offset: offset,
	}
}

func (a *attachments) OnDetached(name string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	delete(a.detaching, name)
}

func (a *attachments) isAttaching(name string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	_, ok := a.attaching[name]
	return ok
}

func (a *attachments) isAttached(name string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	_, ok := a.attached[name]
	return ok
}
