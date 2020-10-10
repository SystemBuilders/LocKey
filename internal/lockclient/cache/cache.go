package cache

// Cache describes an entity of a cache.
type Cache interface {
	// GetElement gets the desired object from the cache.
	// Getting the object makes it the most recently used object
	// in the cache. This function must be implemented in O(1) complexity.
	// If the object doesn't exist in the cache, an error is raised.
	GetElement(element interface{}) (string, error)
	// PutElement inserts an object into the cache.
	// Putting the object makes it the most recently used object
	// in the cache. This function must be implemented in O(1) complexity.
	// If the object already exists in the cache, an error is raised.
	PutElement(element interface{}) error
	// Capacity returns the max capacity of the cache.
	Capacity() int
	// Size returns the number of elements currently in the cache.
	Size() int
	// Full checks whether the cache is full or not. It returns true if the
	// cache is full.
	Full() bool
	// PrintCache prints the cache elements in the decreasing order
	// or use frequency.
	PrintCache()
}
