package cache

// Key descrbes a single key in the Linked List.
type Key interface {
	Data() string
}

// Node describes a single node in the Linked List.
type Node interface {
	Left() Node
	Right() Node
	Key() Key
}

// LinkedList describes a linked list object.
type LinkedList interface {
	// CreateNode creates a node with default values in it.
	CreateNode() Node
	// InsertNodeToLeft inserts a node to the left of the given node,
	// with the key provided. It returns the pointer to the node
	// inserted into the linked list.
	InsertNodeToLeft(node Node, key Key)
	// InsertNodeToRight inserts a node to the right of the given node,
	// with the key provided. It returns the pointer to the node
	// inserted tothe linked list.
	InsertNodeToRight(node Node, key Key)
	// DeleteNode deletes the node provided as the argument from the
	// linked list.
	DeleteNode(Node)
	// PrintLinkedList prints the linked list from head to tail order.
	PrintLinkedList()
}
