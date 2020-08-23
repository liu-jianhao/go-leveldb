package skiplist

type Node struct {
	key  interface{}
	next []*Node
}

func NewNode(key interface{}, height int) *Node {
	return &Node{
		key:  key,
		next: make([]*Node, height),
	}
}

func (n *Node) Next(level int) *Node {
	return n.next[level]
}

func (n *Node) SetNext(level int, node *Node) {
	n.next[level] = node
}
