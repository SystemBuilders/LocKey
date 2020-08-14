package consensus

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

	"github.com/SystemBuilders/LocKey/internal/lockservice"
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
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

// A RaftStore encapsulates the http server (listener),
// a raft node (raftDir, raftAddr, RaftServer) and a
// lock service (SimpleLockService)
type RaftStore struct {
	httpAddr   string
	ls         *lockservice.SimpleLockService
	inmem      bool
	RaftDir    string
	RaftAddr   string
	RaftServer *raft.Raft
	ln         net.Listener
	logger     *log.Logger
}

// New returns a new instance of RaftStore.
func New(inmem bool) *RaftStore {
	return &RaftStore{
		ls:     lockservice.NewSimpleLockService(zerolog.New(os.Stdout).With().Logger().Level(zerolog.GlobalLevel())),
		inmem:  inmem,
		logger: log.New(os.Stderr, "[store] ", log.LstdFlags),
	}
}

// Open opens the store. If enableSingle is set, and there are no existing peers,
// then this node becomes the first node, and therefore leader, of the cluster.
// localID should be the server identifier for this node.
func (rs *RaftStore) Open(enableSingle bool, localID string) error {
	// Setup Raft configuration.
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(localID)

	httpAddr, err := getHTTPAddr(rs.RaftAddr)
	if err != nil {
		return err
	}
	rs.httpAddr = httpAddr

	// Setup Raft communication.
	addr, err := net.ResolveTCPAddr("tcp", rs.RaftAddr)
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(rs.RaftAddr, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(rs.RaftDir, retainSnapshotCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}

	// Create the log store and stable store.
	var logStore raft.LogStore
	var stableStore raft.StableStore
	if rs.inmem {
		logStore = raft.NewInmemStore()
		stableStore = raft.NewInmemStore()
	} else {
		boltDB, err := raftboltdb.NewBoltStore(filepath.Join(rs.RaftDir, "raft.db"))
		if err != nil {
			return fmt.Errorf("new bolt store: %s", err)
		}
		logStore = boltDB
		stableStore = boltDB
	}

	// Instantiate the Raft systems.
	ra, err := raft.NewRaft(config, (*fsm)(rs), logStore, stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	rs.RaftServer = ra

	if enableSingle {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		ra.BootstrapCluster(configuration)
	}

	return nil
}

// Join facilitates the addition of a new Raft node to an existing
// cluster.
//
// The addresss and nodeID are those of the node to be added to the
// cluster.
//
// This function will called on an instance of RaftStore of a node
// existing in the cluster. The function internally sends a HTTP
// request to the listener of the leader of the cluster with the
// new node's ID and address.
func (rs *RaftStore) Join(addr, ID string) error {
	b, err := json.Marshal(map[string]string{"addr": addr, "id": ID})
	if err != nil {
		return err
	}

	var postAddr string
	if rs.RaftServer.Leader() == "" {
		postAddr = rs.RaftAddr
	} else {
		postAddr = string(rs.RaftServer.Leader())
	}

	httpAddr, err := getHTTPAddr(postAddr)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf("http://%s/join", httpAddr),
		"application-type/json",
		bytes.NewReader(b),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	return nil
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
