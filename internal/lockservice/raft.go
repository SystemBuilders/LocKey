package lockservice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/rs/zerolog"
)

const (
	retainSnapshotCount = 2
	raftTimeout         = 10 * time.Second
)

type command struct {
	Op    string `json:"op,omitempty"`
	Key   string `json:"fileID,omitempty"`
	Value string `json:"userID,omitempty"`
}

// A RaftStore encapsulates the http server (listener),
// a raft node (raftDir, raftAddr, RaftServer) and a
// lock service (SimpleLockService)
type RaftStore struct {
	httpAddr   string
	ls         SimpleLockService
	raftDir    string
	raftAddr   string
	RaftServer *raft.Raft
	ln         net.Listener
	logger     *log.Logger
}

// NewRaftServer returns a RaftStore.
func NewRaftServer(raftDir, raftAddr string) (*RaftStore, error) {
	httpAddr, err := getHTTPAddr(raftAddr)
	if err != nil {
		return nil, err
	}
	rs := &RaftStore{
		httpAddr: httpAddr,
		ls:       *NewSimpleLockService(zerolog.New(os.Stdout).With().Logger().Level(zerolog.GlobalLevel())),
		raftDir:  raftDir,
		raftAddr: raftAddr,
		logger:   log.New(os.Stderr, fmt.Sprintf("[raftStore | %s]", raftAddr), log.LstdFlags),
	}

	// what file access controll is this? check.
	if err := os.MkdirAll(raftDir, 0700); err != nil {
		return nil, err
	}

	config := raft.DefaultConfig()
	transport, err := setupRaftCommunication(rs.raftAddr)
	if err != nil {
		return nil, err
	}

	snapshots, err := raft.NewFileSnapshotStore(rs.raftDir, retainSnapshotCount, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("file snapshot store: %s", err)
	}

	boltDB, err := raftboltdb.NewBoltStore(filepath.Join(rs.raftDir, "raft.db"))
	if err != nil {
		return nil, fmt.Errorf("new bolt store: %s", err)
	}
	logStore := boltDB
	stableStore := boltDB

	rft, err := raft.NewRaft(config, (*fsm)(rs), logStore, stableStore, snapshots, transport)

	rs.RaftServer = rft
	return rs, nil
}

func getHTTPAddr(raftAddr string) (string, error) {
	addrParts := strings.Split(raftAddr, ":")
	httpHost := addrParts[0]
	port, err := strconv.Atoi(addrParts[1])
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d", httpHost, port+1), nil

}

func setupRaftCommunication(raftAddr string) (*raft.NetworkTransport, error) {
	addr, err := net.ResolveTCPAddr("tcp", raftAddr)
	if err != nil {
		return nil, err
	}

	// What does the maxPool argument signify?
	transport, err := raft.NewTCPTransport(raftAddr, addr, 3, 10*time.Second, os.Stderr)

	if err != nil {
		return nil, err
	}

	return transport, nil
}

// Acquire locks a File with ID fileID and sets its owner to userID.
// No other user is allowed access to a file once it is locked apart
// from its owner
func (rs *RaftStore) Acquire(fileID, userID string) error {
	b, err := json.Marshal(map[string]string{"fileID": fileID, "userID": userID})
	if err != nil {
		return err
	}

	httpAddr, err := getHTTPAddr(string(rs.RaftServer.Leader()))

	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf("http://%s/acquire", httpAddr),
		"application-type/json",
		bytes.NewReader(b),
	)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil

	// desc := NewSimpleDescriptor(fileID, userID)

	// err := rs.ls.Acquire(desc)
	// if err != nil {
	// 	return err
	// }
	// return nil
}

// // Release calls the lockservice function Release().
// // This in turn checks if userID is the owner of fileID
// // and if it is, fileID is no longer locked.
// // However, if userID does not own fileID, then the lock
// // is not released.
// func (rs *RaftStore) Release(fileID, userID string) error {
// 	b, err := json.Marshal(map[string]string{"fileID": fileID, "userID": userID})
// 	if err != nil {
// 		return err
// 	}

// 	httpAddr, err := getHTTPAddr(string(rs.RaftServer.Leader()))

// 	if err != nil {
// 		return err
// 	}

// 	resp, err := http.Post(
// 		fmt.Sprintf("http://%s/%s/fileID/%s", httpAddr, rs.raftDir, fileID),
// 		"application-type/json",
// 		bytes.NewReader(b),
// 	)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	return nil

// 	// desc := NewSimpleDescriptor(fileID, userID)

// 	// err := rs.ls.Release(desc)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// return nil
// }
