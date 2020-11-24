package lockclient

import (
	"github.com/SystemBuilders/LocKey/internal/lockclient/session"
	"github.com/SystemBuilders/LocKey/internal/lockservice"
)

// Client describes a client that can be used to interact with
// the Lockey lockservice. The client can start the lockservice
// and interact by making calls to it.
//
// The client has the ability to start the lockservice from its
// in-built function or it can be started separately.
//
// The client allows the user to Acquire a lock and Release a lock,
// using it's descriptor.
type Client interface {
	// StartService starts the lockservice Lockey using the given
	// configuration. It provides an appropriate error on failing
	// to do so. Starting the service should be a non-blocking call
	// and return as soon as the server is started and setup.
	StartService(Config) error
	// Connect allows the user process to establish a connection
	// with the client. This returns an ID of the session that
	// results from the connection.
	Connect() session.Session
	// Acquire can be used to acquire a lock on Lockey. This
	// implementation interacts with the underlying server and
	// provides the service.
	Acquire(lockservice.Object, session.Session) error
	// Release can be used to release a lock on Lockey. This
	// implementation interacts with the underlying server and
	// provides the service.
	Release(lockservice.Object, session.Session) error
}

// Config describes the configuration for the lockservice to run on.
type Config interface {
	// IP provides the IP address where the server is intended to run.
	IP() string
	// Port provides the port where the server is supposed to run.
	Port() string
}
