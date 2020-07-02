package cache

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
type LRUCache struct {
	capacity int
	size     int
	full     bool
	m        map[interface{}]bool
	dll      *DoublyLinkedList
}

// NewLRUCache creates a new LRUCache of provided size.
func NewLRUCache(size int) *LRUCache {
	return &LRUCache{
		capacity: 0,
		size:     size,
		full:     false,
		m:        make(map[interface{}]bool),
		dll:      NewDoublyLinkedList(),
	}
}

// GetElement gets an element from the cache.
func (lru *LRUCache) GetElement(element interface{}) error {
	panic("TODO")
}

// PutElement inserts an element in the cache.
func (lru *LRUCache) PutElement(element interface{}) error {
	panic("TODO")
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
	panic("TODO")
}
