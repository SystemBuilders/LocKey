package cache

import (
	"fmt"
	"log"
)

// Assert that *SimpleKey implements Key.
var _ Key = (*SimpleKey)(nil)

// SimpleKey implements a Key interface.
type SimpleKey struct {
	Owner string
	Value string
}

// Data returns the value of the key.
func (sk *SimpleKey) Data() string {
	return sk.Value
}

// NewSimpleKey returns a new SimpleKey of the given value.
func NewSimpleKey(val, owner string) *SimpleKey {
	return &SimpleKey{
		Value: val,
		Owner: owner,
	}
}

// Assert that *DLLNode implements Node.
var _ Node = (*DLLNode)(nil)

// Left returns the node to the left of the current node.
func (dllNode *DLLNode) Left() Node {
	return dllNode.LeftNode
}

// Right returns the node to the right of the current node.
func (dllNode *DLLNode) Right() Node {
	return dllNode.RightNode
}

// Key returns the key of the node.
func (dllNode *DLLNode) Key() Key {
	return dllNode.NodeKey
}

// DLLNode is the single entity of the doubly linked list.
type DLLNode struct {
	LeftNode  Node
	RightNode Node
	NodeKey   *SimpleKey
}

// Assert that *DoublyLinkedList implements LinkedList.
var _ LinkedList = (*DoublyLinkedList)(nil)

// DoublyLinkedList implements LinkedList.
//
// All nodes have a up and a down link except the head and the tail node.
type DoublyLinkedList struct {
	Head Node
}

// NewDoublyLinkedList returns a new instance of an empty DoublyLinkedList.
func NewDoublyLinkedList() *DoublyLinkedList {
	return &DoublyLinkedList{
		Head: nil,
	}
}

// CreateNode creates an empty node of the DLL with default values.
func (dll *DoublyLinkedList) CreateNode() Node {
	return &DLLNode{
		LeftNode:  nil,
		RightNode: nil,
		NodeKey: &SimpleKey{
			Value: "",
		},
	}
}

// InsertNodeToLeft inserts a node with given value to the left of the given node.
func (dll *DoublyLinkedList) InsertNodeToLeft(node Node, key Key) {
	if node == nil || node.(*DLLNode) == nil {
		// If the node to be inserted on is nil, assign the newNode to the same and return.
		newNode := dll.CreateNode()
		newNode.(*DLLNode).NodeKey = key.(*SimpleKey)
		// Since the DLL was empty before, this will be the Head node too.
		dll.Head = newNode
		return
	}
	if node.Left() == nil {
		newNode := dll.CreateNode()
		newNode.(*DLLNode).NodeKey = key.(*SimpleKey)
		node.(*DLLNode).LeftNode = newNode
		newNode.(*DLLNode).RightNode = node
		if node.(*DLLNode) == dll.Head {
			dll.Head = newNode
		}
		return
	}
	newNode := dll.CreateNode()
	newNode.(*DLLNode).NodeKey = key.(*SimpleKey)
	leftNode := node.Left()

	node.(*DLLNode).LeftNode = newNode
	newNode.(*DLLNode).LeftNode = leftNode

	leftNode.(*DLLNode).RightNode = newNode
	newNode.(*DLLNode).RightNode = node
}

// InsertNodeToRight inserts a node with the given value to the right of the given node.
func (dll *DoublyLinkedList) InsertNodeToRight(node Node, key Key) {
	if node == nil || node.(*DLLNode) == nil {
		newNode := dll.CreateNode()
		newNode.(*DLLNode).NodeKey = key.(*SimpleKey)
		dll.Head = newNode
		return
	}
	if node.Right() == nil {
		newNode := dll.CreateNode()
		newNode.(*DLLNode).NodeKey = key.(*SimpleKey)
		node.(*DLLNode).RightNode = newNode
		newNode.(*DLLNode).LeftNode = node
		return
	}
	newNode := dll.CreateNode()
	newNode.(*DLLNode).NodeKey = key.(*SimpleKey)

	rightNode := node.Right()

	node.(*DLLNode).RightNode = newNode
	newNode.(*DLLNode).RightNode = rightNode

	rightNode.(*DLLNode).LeftNode = newNode
	newNode.(*DLLNode).LeftNode = node
}

// DeleteNode deletes the provided node.
func (dll *DoublyLinkedList) DeleteNode(node Node) {
	if node == nil {
		log.Println("Node is nil, can't be deleted")
	}

	leftNode := node.Left()
	rightNode := node.Right()

	if node == dll.Head {
		dll.Head = rightNode
	}

	if leftNode != nil {
		leftNode.(*DLLNode).RightNode = rightNode
		if rightNode != nil {
			rightNode.(*DLLNode).LeftNode = leftNode
		}
	} else {
		if rightNode != nil {
			rightNode.(*DLLNode).LeftNode = nil
		}
	}
}

// PrintLinkedList prints the given linked list from head to tail order.
func (dll *DoublyLinkedList) PrintLinkedList() {
	head := dll.Head
	for {
		if head == nil {
			break
		}
		fmt.Printf("%d-> ", head.Key())
		head = head.(*DLLNode).RightNode
	}
	fmt.Println("NULL")
}
