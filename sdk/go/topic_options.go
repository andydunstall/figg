package figg

type TopicOptions struct {
	// Offset is an offset of an old message to subscribe from without missing
	// messages. This is only used if FromOffset is true, otherwise is ignored.
	Offset     uint64
	FromOffset bool
}

type TopicOption func(*TopicOptions)

func WithOffset(offset uint64) TopicOption {
	return func(opts *TopicOptions) {
		opts.Offset = offset
		opts.FromOffset = true
	}
}

func defaultTopicOptions() *TopicOptions {
	return &TopicOptions{
		Offset:     0,
		FromOffset: false,
	}
}
