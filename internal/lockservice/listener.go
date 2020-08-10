package lockservice

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

func (rs *RaftStore) Start() error {
	server := http.Server{
		Handler: rs,
	}

	ln, err := net.Listen("tcp", rs.httpAddr)
	if err != nil {
		return err
	}
	rs.ln = ln

	http.Handle(fmt.Sprintf("/%s", rs.raftDir), rs)

	fmt.Printf("set up listener at %s\n", rs.httpAddr)
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
