package lockclient

import "github.com/GoPlayAndFun/LocKey/internal/lockservice"

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
}

// Config describes the configuration for the lockservice to run on.
type Config interface {
	// IP provides the IP address where the server is intended to run.
	IP() string
	// Port provides the port where the server is supposed to run.
	Port() string
}
