package lockservice

import (
	"log"
	"net"
	"net/http"
)

// Start starts the http service using the listener within a RaftStore.
// The HTTP server is used to redirect commands like Set, Delete and Join
// to the leader RaftStore in a cluster. The HTTP address is always
// one away from the Raft address which the raft node uses for communication
// with other raft nodes.
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

func (rs *RaftStore) Close() {
	rs.ln.Close()
	return
}
