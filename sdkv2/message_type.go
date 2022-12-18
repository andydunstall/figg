package figg

type MessageType uint16

const (
	TypeAttach   = MessageType(1)
	TypeAttached = MessageType(2)
)

func (t MessageType) String() string {
	switch t {
	case TypeAttach:
		return "ATTACH"
	case TypeAttached:
		return "ATTACHED"
	default:
		return "UNKNOWN"
	}
}
