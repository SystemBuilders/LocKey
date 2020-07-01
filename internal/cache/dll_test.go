package cache

import "testing"

func Test_DLL(t *testing.T) {
	dllNode := NewDoublyLinkedList()

	dllNode.InsertNodeToRight(nil, &SimpleKey{10})

	if dllNode.Head.Key().Data() != 10 {
		t.Errorf("Required value 10, got %d", dllNode.Head.Key())
	}

	dllNode.InsertNodeToRight(dllNode.Head, &SimpleKey{20})
	dllNode.InsertNodeToRight(dllNode.Head, &SimpleKey{30})
	dllNode.InsertNodeToRight(dllNode.Head, &SimpleKey{40})
	dllNode.InsertNodeToRight(dllNode.Head, &SimpleKey{50})

	dllNode.PrintLinkedList()

	newNode := dllNode.Head.Right().Right()

	dllNode.InsertNodeToLeft(newNode, &SimpleKey{60})
	dllNode.InsertNodeToLeft(newNode, &SimpleKey{70})

	dllNode.PrintLinkedList()

	dllNode.DeleteNode(dllNode.Head)
	dllNode.DeleteNode(dllNode.Head)

	dllNode.PrintLinkedList()

	dllNode.InsertNodeToRight(dllNode.Head, &SimpleKey{80})
	dllNode.InsertNodeToLeft(dllNode.Head, &SimpleKey{90})

	dllNode.PrintLinkedList()
}
