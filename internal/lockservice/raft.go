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
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

// A RaftStore encapsulates the http server (listener),
// a raft node (raftDir, raftAddr, RaftServer) and a
// lock service (SimpleLockService)
type RaftStore struct {
	httpAddr   string
	ls         *SimpleLockService
	inmem      bool
	RaftDir    string
	RaftAddr   string
	RaftServer *raft.Raft
	ln         net.Listener
	logger     *log.Logger
}

// New returns a new Store.
func New(inmem bool) *RaftStore {
	return &RaftStore{
		ls:     NewSimpleLockService(zerolog.New(os.Stdout).With().Logger().Level(zerolog.GlobalLevel())),
		inmem:  inmem,
		logger: log.New(os.Stderr, "[store] ", log.LstdFlags),
	}
}

// Open opens the store. If enableSingle is set, and there are no existing peers,
// then this node becomes the first node, and therefore leader, of the cluster.
// localID should be the server identifier for this node.
func (s *RaftStore) Open(enableSingle bool, localID string) error {
	// Setup Raft configuration.
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(localID)
	// config.ProtocolVersion = 2

	httpAddr, err := getHTTPAddr(s.RaftAddr)
	if err != nil {
		return err
	}
	s.httpAddr = httpAddr

	// Setup Raft communication.
	addr, err := net.ResolveTCPAddr("tcp", s.RaftAddr)
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(s.RaftAddr, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return err
	}

	// Create the snapshot store. This allows the Raft to truncate the log.
	snapshots, err := raft.NewFileSnapshotStore(s.RaftDir, retainSnapshotCount, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}

	// Create the log store and stable store.
	var logStore raft.LogStore
	var stableStore raft.StableStore
	if s.inmem {
		logStore = raft.NewInmemStore()
		stableStore = raft.NewInmemStore()
	} else {
		boltDB, err := raftboltdb.NewBoltStore(filepath.Join(s.RaftDir, "raft.db"))
		if err != nil {
			return fmt.Errorf("new bolt store: %s", err)
		}
		logStore = boltDB
		stableStore = boltDB
	}

	// Instantiate the Raft systems.
	ra, err := raft.NewRaft(config, (*fsm)(s), logStore, stableStore, snapshots, transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	s.RaftServer = ra

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
	defer resp.Body.Close()
	if err != nil {
		return err
	}

	return nil
}

// // Join joins a node, identified by nodeID and located at addr, to this store.
// // The node must be ready to respond to Raft communications at that address.
// func (s *RaftStore) Join(nodeID, addr string) error {
// 	s.logger.Printf("received join request for remote node %s at %s", nodeID, addr)

// 	configFuture := s.RaftServer.GetConfiguration()
// 	if err := configFuture.Error(); err != nil {
// 		s.logger.Printf("failed to get raft configuration: %v", err)
// 		return err
// 	}

// 	for _, srv := range configFuture.Configuration().Servers {
// 		// If a node already exists with either the joining node's ID or address,
// 		// that node may need to be removed from the config first.
// 		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(addr) {
// 			// However if *both* the ID and the address are the same, then nothing -- not even
// 			// a join operation -- is needed.
// 			if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(nodeID) {
// 				s.logger.Printf("node %s at %s already member of cluster, ignoring join request", nodeID, addr)
// 				return nil
// 			}

// 			future := s.RaftServer.RemoveServer(srv.ID, 0, 0)
// 			if err := future.Error(); err != nil {
// 				return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, addr, err)
// 			}
// 		}
// 	}

// 	f := s.RaftServer.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
// 	if f.Error() != nil {
// 		return f.Error()
// 	}
// 	s.logger.Printf("node %s at %s joined successfully", nodeID, addr)
// 	return nil
// }

func getHTTPAddr(raftAddr string) (string, error) {
	addrParts := strings.Split(raftAddr, ":")
	httpHost := addrParts[0]
	port, err := strconv.Atoi(addrParts[1])
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d", httpHost, port+1), nil

}
