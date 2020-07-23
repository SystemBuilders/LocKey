package lockclient

import "github.com/SystemBuilders/LocKey/internal/lockservice"

// Client describes a client that can be used to interact with
// the Lockey lockservice. The client can start the lockservice
// and interact acquire and release locks by making calls to it.
type Client interface {
	// StartService starts the lockservice Lockey using the given
	// configuration. It provides an appropriate error on failing
	// to do so. Starting the service should be a non-blocking call
	// and return as soon as the server is started and setup.
	StartService(Config) error
	// Acquire can be used to acquire a lock on Lockey. This
	// implementation interacts with the underlying server and
	// provides the service.
	Acquire(lockservice.Descriptors) error
	// Release can be used to release a lock on Lockey. This
	// implementation interacts with the underlying server and
	// provides the service.
	Release(lockservice.Descriptors) error
	// Watch can be used to watch the state of lock on a descriptor
	// continously. When the state of the lock changes, the "watcher"
	// will be notified about the change.
	// The channel passed as the argument can be used to stop watching
	// at any point of time.
	Watch(lockservice.Descriptors, chan struct{}) (chan Lock, error)
	// Pounce can be used to "pounce" on a lock that has already been
	// acquired. This is similar to acquire but once a process has
	// opted to pounce, they will be provided first access by having
	// a queue of pouncers.
	Pounce(lockservice.Descriptors) error
	// Pouncers returns the current pouncers on any particular lock.
	Pouncers(lockservice.Descriptors) []string
}

// Config describes the configuration for the lockservice to run on.
type Config interface {
	// IP provides the IP address where the server is intended to run.
	IP() string
	// Port provides the port where the server is supposed to run.
	Port() string
}

// State describes the state of a lock.
type State int

// These are the states of a lock.
const (
	Acquire State = iota
	Release
)

// Lock includes the state of the lock and the owner.
type Lock struct {
	Owner     string
	LockState State
}
