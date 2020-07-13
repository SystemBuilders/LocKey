package cache

import "fmt"

var _ Cache = (*LRUCache)(nil)

// LRUCache implements a cache. It uses a linked list as
// the primary data structure along with a hash-map for
// checking existance of an element in the cache.
//
// The starting element in the linked list will always be
// the most recently used element in the cache and will be
// maintained that way by all the operating functions.
//
// GetElement and PutElement inherently implement a method
// to rank the elements on the basis of frequency.
// This frequency based order is controlled by:
// * Maintaining a logical order in the DLL - first element is MRU.
// * At every insertion, the MRU is maintained at the Head of the DLL.
// * After every access, the element is moved to the MRU position in the DLL.
// * All insertions happen at the "start" pointer for ease and this
//   is repositioned after every access to the appropriate next position.
// * The "start" pointer ALWAYS points to the LRU element and
//   order of the DLL is decreasing frequency of the elements.
//
// The hash map maintains the existance of the element in the cache
// and the DLL is to maintain the frequency of the usage of the element.
type LRUCache struct {
	capacity int
	size     int
	full     bool
	tail     *DLLNode
	m        map[interface{}]*DLLNode
	dll      *DoublyLinkedList
}

// NewLRUCache creates a new LRUCache of provided size.
func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		size:     0,
		full:     false,
		m:        make(map[interface{}]*DLLNode),
		dll:      NewDoublyLinkedList(),
	}
}

// GetElement gets an element from the cache.
//
// Whenever an element is retrieved from the cache,
// it's bumped to the MRU position in the DLL.
// The "start" pointer is always positioned to the
// "next addition ready" position.
func (lru *LRUCache) GetElement(element interface{}) error {
	// Check whether the element exists in the cache.
	if _, ok := lru.m[element]; ok {
		nodeOfKey := lru.m[element]
		// Check whether the currently accessed element is the
		// most recently used element in the cache. If it's not,
		// it must be moved to the MRU to accomodate the protocol.
		if lru.dll.Head.Key() != element && lru.tail != lru.dll.Head {
			fmt.Println(nodeOfKey.Key())
			if lru.tail.Key() == element {
				lru.tail = lru.tail.LeftNode.(*DLLNode)
			}
			leftNode := nodeOfKey.LeftNode
			rightNode := nodeOfKey.RightNode
			lru.dll.DeleteNode(nodeOfKey)
			if leftNode != nil {
				lru.insertElementIntoMap(leftNode.Key(), leftNode)
			}
			if rightNode != nil {
				lru.insertElementIntoMap(rightNode.Key(), rightNode)
			}
			// Move the currently accessed node to the MRU position.
			// The start pointer doesn't change as it still points to
			// the LRU element.
			lru.dll.InsertNodeToLeft(lru.dll.Head, nodeOfKey.NodeKey)
			lru.insertElementIntoMap(element, lru.dll.Head)
			headRight := lru.dll.Head.Right()
			if headRight != nil {
				lru.insertElementIntoMap(headRight.Key(), headRight)
			}
		}
		return nil
	}
	return ErrElementDoesntExist
}

// PutElement inserts an element in the cache.
func (lru *LRUCache) PutElement(element interface{}) error {
	if !lru.full {
		if lru.dll.Head == nil {
			lru.dll.InsertNodeToRight(lru.dll.Head, element.(*SimpleKey))
			err := lru.insertElementIntoMap(element, lru.dll.Head)
			if err != nil {
				return err
			}
			lru.tail = lru.dll.Head.(*DLLNode)
		} else {
			lru.dll.InsertNodeToLeft(lru.dll.Head, element.(*SimpleKey))
			err := lru.insertElementIntoMap(element, lru.dll.Head)
			if err != nil {
				return err
			}
		}
		lru.size++
		if lru.size == lru.capacity {
			lru.full = true
		}
	} else {
		lru.dll.InsertNodeToLeft(lru.dll.Head, element.(*SimpleKey))
		err := lru.insertElementIntoMap(element, lru.dll.Head)
		if err != nil {
			return err
		}

		tailNode := lru.tail
		lru.tail = lru.tail.LeftNode.(*DLLNode)
		// Delete the "start" node and make the newly inserted node the MRU node.
		lru.dll.DeleteNode(tailNode)
		delete(lru.m, tailNode.Key())
	}
	return nil
}

// Capacity returns the max capacity of the cache.
func (lru *LRUCache) Capacity() int {
	return lru.capacity
}

// Size returns the number of elements in the cache.
func (lru *LRUCache) Size() int {
	return lru.size
}

// Full returns true if the cache is full, else returns false.
func (lru *LRUCache) Full() bool {
	return lru.full
}

// PrintCache prints the entire cache in decreasing order of frequency of usage.
func (lru *LRUCache) PrintCache() {
	fmt.Println("Cache:")
	fmt.Printf("\nMap:\n")
	lru.printMap()
	fmt.Printf("\nDLL:\n")
	lru.dll.PrintLinkedList()
	if lru.tail != nil {
		fmt.Printf("\n\nTail: %d\nFull? : %t\n", lru.tail.Key(), lru.full)
	} else {
		fmt.Printf("\n\nTail: %d\nFull? : %t\n", -1, lru.full)
	}
	fmt.Printf("\n\n-------------------------------------\n\n")
}

func (lru *LRUCache) insertElementIntoMap(key interface{}, node Node) error {
	if _, ok := lru.m[key]; !ok {
		lru.m[key] = node.(*DLLNode)
	} else {
		return ErrElementAlreadyExists
	}
	return nil
}

func (lru *LRUCache) printMap() {
	for k, v := range lru.m {
		fmt.Printf("Key: %v, Value: ", k)
		fmt.Printf("LN: %v, RN: %v, NodeKey: %v\n", v.Left(), v.Right(), v.Key())
	}
}
