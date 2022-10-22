package topic

type Subscriber interface {
	Notify(b []byte)
}
