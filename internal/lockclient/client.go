package lockclient

import "github.com/SystemBuilders/LocKey/internal/lockservice"

// Client describes a client that can be used to interact with
// the Lockey lockservice. The client can start the lockservice
// and interact by making calls to it.
//
// The client has the ability to start the lockservice from its
// in-built function or it can be started separately.
//
// The client offers the user to Acquire a lock, Release a lock,
// and Watch or Pounce on any object using it's descriptor.
//
// To acquire a lock on an object, the user is forced to go via
// the Pounce function in order to maintain the order of lock
// acquisition. The Acquire function that is exposed must cleverly
// handle this problem.
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
	watch(lockservice.ObjectDescriptor, chan struct{}) (chan Lock, error)
	// Pounce can be used to "pounce" on a lock that has already been
	// acquired. This is similar to acquire but once a process has
	// opted to pounce, they will be provided first access by having
	// a queue of pouncers.
	// The second, third and fourth arguments dictate the end of the pouncing
	// reign, the owner willing to pounce and allows pouncing on pre-pounced
	// objects respectively. True bool allows pouncing on pre-pounced objects.
	pounce(lockservice.ObjectDescriptor, string, chan struct{}, bool) error
	// Pouncers returns the current pouncers on any particular lock.
	pouncers(lockservice.ObjectDescriptor) []string
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
	acquire State = iota
	Release
)

// Lock includes the state of the lock and the owner.
type Lock struct {
	Owner     string
	LockState State
}
