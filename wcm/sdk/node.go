package wcm

type Node struct {
	ID   string
	Addr string
}

func NewNode(id string, addr string) *Node {
	return &Node{
		ID:   id,
		Addr: addr,
	}
}
