package figg

type MessageType uint16

const (
	TypeAttach   = MessageType(1)
	TypeAttached = MessageType(2)
	TypeDetach   = MessageType(3)
	TypeDetached = MessageType(4)
)

func (t MessageType) String() string {
	switch t {
	case TypeAttach:
		return "ATTACH"
	case TypeAttached:
		return "ATTACHED"
	case TypeDetach:
		return "DETACH"
	case TypeDetached:
		return "DETACHED"
	default:
		return "UNKNOWN"
	}
}
