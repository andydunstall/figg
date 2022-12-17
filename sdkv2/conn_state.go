package figg

type ConnState int

const (
	DISCONNECTED = ConnState(iota)
	CONNECTED
)

func (c ConnState) String() string {
	switch c {
	case DISCONNECTED:
		return "DISCONNECTED"
	case CONNECTED:
		return "CONNECTED"
	default:
		return "UNKNOWN"
	}
}
