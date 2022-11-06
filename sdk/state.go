package wombat

type State int

const (
	StateConnected    = State(1)
	StateDisconnected = State(2)
)

func StateToString(s State) string {
	switch s {
	case StateConnected:
		return "CONNECTED"
	case StateDisconnected:
		return "DISCONNECTED"
	default:
		return "UNKNOWN"
	}
}
