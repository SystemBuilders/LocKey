package cache

import (
	"fmt"
	"sync"
)

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
// * All insertions occur at the head of the DLL since this is the
//   MRU position. This ensures that the LRU position is the tail.
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
	mu       sync.Mutex
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

// GetElement gets an element from the cache. It returns
// the associated data with the element with an error.
//
// Whenever an element is retrieved from the cache,
// it's bumped to the MRU position in the DLL.
//
// The element is removed from the map too because
// it might have stale node values.
//
// Error is returned only if the element doesn't exist in the cache.
func (lru *LRUCache) GetElement(element interface{}) (string, error) {
	// Check whether the element exists in the cache.
	lru.mu.Lock()
	defer lru.mu.Unlock()
	if node, ok := lru.m[*&element.(*SimpleKey).Value]; ok {
		nodeOfKey := lru.m[*&element.(*SimpleKey).Value]
		// Check whether the currently accessed element is the
		// most recently used element in the cache. If it's not,
		// it must be moved to the MRU to accomodate the protocol.
		if lru.dll.Head.Key() != element {
			// update the tail node.
			if lru.tail.Key() == element {
				lru.tail = lru.tail.LeftNode.(*DLLNode)
			}

			// delete the old key so that it can be moved to MRU.
			lru.dll.DeleteNode(nodeOfKey)
			lru.deleteElementFromMap(*&nodeOfKey.Key().(*SimpleKey).Value)

			// Move the currently accessed node to the MRU position.
			// The start pointer doesn't change as it still points to
			// the LRU element.
			lru.dll.InsertNodeToLeft(lru.dll.Head, nodeOfKey.NodeKey)
			lru.insertElementIntoMap(*&element.(*SimpleKey).Value, lru.dll.Head)
		}
		return node.NodeKey.Owner, nil
	}
	return "", ErrElementDoesntExist
}

// PutElement inserts an element in the cache.
// All insertions occur at the head node of the DLL.
//
// Removal of the LRU is done my deleting the tail node,
// making place for a new node.
func (lru *LRUCache) PutElement(element interface{}) error {
	lru.mu.Lock()
	defer lru.mu.Unlock()
	if !lru.full {
		if lru.dll.Head == nil {
			lru.dll.InsertNodeToRight(lru.dll.Head, element.(*SimpleKey))
			err := lru.insertElementIntoMap(*&element.(*SimpleKey).Value, lru.dll.Head)
			if err != nil {
				return err
			}
			lru.tail = lru.dll.Head.(*DLLNode)
		} else {
			lru.dll.InsertNodeToLeft(lru.dll.Head, element.(*SimpleKey))
			err := lru.insertElementIntoMap(*&element.(*SimpleKey).Value, lru.dll.Head)
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
		err := lru.insertElementIntoMap(*&element.(*SimpleKey).Value, lru.dll.Head)
		if err != nil {
			return err
		}

		tailNode := lru.tail
		lru.tail = lru.tail.LeftNode.(*DLLNode)

		// Delete the "start" node and make the newly inserted node the MRU node.
		lru.dll.DeleteNode(tailNode)
		lru.deleteElementFromMap(*&tailNode.Key().(*SimpleKey).Value)
	}
	return nil
}

// RemoveElement deletes a node from the cache based on a key value
// If there are multiple nodes with the same value, the node that was
// most recently used will be removed.
func (lru *LRUCache) RemoveElement(element interface{}) error {
	lru.mu.Lock()
	defer lru.mu.Unlock()
	// Check if the node exists in the cache
	if _, ok := lru.m[*&element.(*SimpleKey).Value]; ok {
		nodeOfKey := lru.m[*&element.(*SimpleKey).Value]
		// If there is only one element in the linked list, make the
		// tail point to nil
		//
		// If the element being deleted is the tail, change the tail of
		// the linked list to its left node
		if lru.tail == lru.dll.Head {
			lru.tail = nil
		} else if lru.tail == nodeOfKey {
			lru.tail = nodeOfKey.LeftNode.(*DLLNode)
		}
		lru.size--
		lru.dll.DeleteNode(nodeOfKey)
		lru.deleteElementFromMap(*&nodeOfKey.Key().(*SimpleKey).Value)
		return nil

	}
	return ErrElementDoesntExist
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
	if lru.tail != nil && lru.dll.Head != nil {
		fmt.Printf("\n\nHead: %s\nTail: %s\nFull? : %t\n", lru.dll.Head.Key(), lru.tail.Key(), lru.full)
	} else if lru.tail == nil && lru.dll.Head != nil {
		fmt.Printf("\n\nHead: %s\nTail: %s\nFull? : %t\n", lru.dll.Head.Key(), "", lru.full)
	} else {
		fmt.Printf("\n\nHead: %s\nTail: %s\nFull? : %t\n", "", "", lru.full)
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

func (lru *LRUCache) deleteElementFromMap(key interface{}) error {
	if _, ok := lru.m[key]; ok {
		delete(lru.m, key)
	} else {
		return ErrElementDoesntExist
	}
	return nil
}
func (lru *LRUCache) printMap() {
	for k, v := range lru.m {
		fmt.Printf("Key: %v, Value: ", k)
		fmt.Printf("LN: %v, RN: %v, NodeKey: %v\n", v.Left(), v.Right(), v.Key())
	}
}
