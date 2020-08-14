package consensus

import (
	"log"
	"net"
	"net/http"
)

// Start starts the http service using the listener within a RaftStore.
//
// The HTTP server is used to redirect commands like Set, Delete and Join
// to the leader RaftStore in a cluster. The HTTP address is always
// one away from the Raft address which the raft node uses for communication
// with other raft nodes.
//
// This policy is maybe a bit to trivial and in the future, a more dynamic
// mapping between a Raft node and its listener can be integrated
func (rs *RaftStore) Start() error {
	server := http.Server{
		Handler: rs,
	}

	ln, err := net.Listen("tcp", rs.httpAddr)
	if err != nil {
		return err
	}
	rs.ln = ln

	go func() {
		err := server.Serve(rs.ln)
		if err != nil {
			log.Fatalf("HTTP serve: %s", err)
		}
	}()

	return nil
}

// Close stops the listener corresponding to a Raft node.
func (rs *RaftStore) Close() {
	rs.ln.Close()
	return
}
