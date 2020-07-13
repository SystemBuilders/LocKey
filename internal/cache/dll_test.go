package cache

import "testing"

func Test_DLL(t *testing.T) {
	dllNode := NewDoublyLinkedList()

	dllNode.InsertNodeToRight(nil, &SimpleKey{1})
	dllNode.PrintLinkedList()

	if dllNode.Head.Key().Data() != 1 {
		t.Errorf("Required value 1, got %d", dllNode.Head.Key())
	}

	dllNode.InsertNodeToLeft(dllNode.Head, &SimpleKey{2})
	dllNode.PrintLinkedList()

	dllNode.DeleteNode(dllNode.Head.Right())
	dllNode.PrintLinkedList()

	dllNode.InsertNodeToLeft(dllNode.Head, &SimpleKey{1})
	dllNode.PrintLinkedList()

	dllNode.InsertNodeToLeft(dllNode.Head, &SimpleKey{3})
	dllNode.PrintLinkedList()

	dllNode.DeleteNode(dllNode.Head.Right().Right())
	dllNode.PrintLinkedList()

	dllNode.InsertNodeToLeft(dllNode.Head, &SimpleKey{2})
	dllNode.PrintLinkedList()

	dllNode.InsertNodeToLeft(dllNode.Head, &SimpleKey{4})
	dllNode.PrintLinkedList()

	dllNode.DeleteNode(dllNode.Head.Right().Right().Right())
	dllNode.PrintLinkedList()

	dllNode.InsertNodeToLeft(dllNode.Head, &SimpleKey{1})
	dllNode.PrintLinkedList()
}
