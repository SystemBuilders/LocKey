package lockservice

// LockService describes a lock service component that enables
// maintaining a set of locks. This service is a standalone component
// that can be implemented on any server component, distributed or not.
type LockService interface {
	// Acquire allows the service to set a lock on the given descriptors.
	// An error is generated if the same isn't possible for any reason,
	// including already existing locks on the descriptor.
	Acquire(Descriptors) error
	// Release allows the service to release the lock on the given descriptors.
	// An error is generated if the same isn't possible for any reason,
	// including releasing locks on non-acquired descriptors.
	Release(Descriptors) error
	// CheckAcquire checks whether a lock has been acquired on the given descriptor.
	// The function returns true if the lock has been acquired on the component.
	CheckAcquired(Descriptors) bool
	// CheckRelease checks whether a lock has been released (or not acquired) on the
	// given component. Returns true if there are no locks on the descriptor.
	CheckReleased(Descriptors) bool
}

// Descriptors describe the type of data that a lock acquiring component must describe.
type Descriptors interface {
	ID() string
}
