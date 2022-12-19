package utils

type MessageType uint16

const (
	TypeAttach   = MessageType(1)
	TypeAttached = MessageType(2)
	TypeDetach   = MessageType(3)
	TypeDetached = MessageType(4)
	TypePublish  = MessageType(5)
	TypeACK      = MessageType(6)
	TypeData     = MessageType(7)
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
	case TypePublish:
		return "PUBLISH"
	case TypeACK:
		return "ACK"
	case TypeData:
		return "DATA"
	default:
		return "UNKNOWN"
	}
}
