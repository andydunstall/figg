package figg

type Figg struct {
}

// Connect will attempt to connect to the given Figg node.
func Connect(addr string, options ...Option) (*Figg, error) {
	opts := defaultOptions(addr)
	for _, opt := range options {
		opt(opts)
	}

	_, err := opts.Dialer.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Figg{}, nil
}
