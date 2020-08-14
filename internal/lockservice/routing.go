package lockservice

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/raft"
)

func (rs *RaftStore) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if getRaftAddr(rs.httpAddr) != string(rs.RaftServer.Leader()) {
		url := r.URL
		url.Host, _ = getHTTPAddr(string(rs.RaftServer.Leader()))
		url.Scheme = "http"

		proxyReq, err := http.NewRequest(r.Method, url.String(), r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		proxyReq.Header.Set("Host", r.Host)
		proxyReq.Header.Set("X-Forwarded-For", r.RemoteAddr)

		for header, values := range r.Header {
			for _, value := range values {
				proxyReq.Header.Add(header, value)
			}
		}

		client := &http.Client{}
		resp, err := client.Do(proxyReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		return
	}
	if strings.Contains(r.URL.Path, "/acquire") {
		rs.handleAcquire(w, r)
	} else if strings.Contains(r.URL.Path, "/release") {
		rs.handleRelease(w, r)
	} else if strings.Contains(r.URL.Path, "/join") {
		rs.handleJoin(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (rs *RaftStore) handleAcquire(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req LockRequest
	err = json.Unmarshal(body, &req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c := &command{
		Op:    "acquire",
		Key:   req.FileID,
		Value: req.UserID,
	}
	b, err := json.Marshal(c)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if acquire is possible
	desc := NewSimpleDescriptor(req.FileID, req.UserID)
	err = rs.ls.TryAcquire(desc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// If possible, commit the change
	f := rs.RaftServer.Apply(b, raftTimeout)
	if f.Error() != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Write([]byte("lock acquired"))
}

func (rs *RaftStore) handleRelease(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req LockRequest
	err = json.Unmarshal(body, &req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	c := &command{
		Op:    "release",
		Key:   req.FileID,
		Value: req.UserID,
	}
	b, err := json.Marshal(c)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if release is possible
	desc := NewSimpleDescriptor(req.FileID, req.UserID)
	err = rs.ls.TryRelease(desc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// If possible, commit the change
	f := rs.RaftServer.Apply(b, raftTimeout)
	if f.Error() != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Write([]byte("lock released"))
}

// handleJoin actually applies the join upon receiving the http request.
func (rs *RaftStore) handleJoin(w http.ResponseWriter, r *http.Request) {
	m := map[string]string{}
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	remoteAddr := m["addr"]
	nodeID := m["id"]

	// f := rs.RaftServer.AddPeer(raft.ServerAddress(remoteAddr))
	err = rs.joinHelper(nodeID, remoteAddr)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rs.logger.Printf("node at %s joined successfully", remoteAddr)
	w.Write([]byte("joined cluster"))
}

// Join joins a node, identified by nodeID and located at addr, to this store.
// The node must be ready to respond to Raft communications at that address.
func (rs *RaftStore) joinHelper(nodeID, addr string) error {
	rs.logger.Printf("received join request for remote node %s at %s", nodeID, addr)

	configFuture := rs.RaftServer.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		rs.logger.Printf("failed to get raft configuration: %v", err)
		return err
	}

	for _, srv := range configFuture.Configuration().Servers {
		// If a node already exists with either the joining node's ID or address,
		// that node may need to be removed from the config first.
		if srv.ID == raft.ServerID(nodeID) || srv.Address == raft.ServerAddress(addr) {
			// However if *both* the ID and the address are the same, then nothing -- not even
			// a join operation -- is needed.
			if srv.Address == raft.ServerAddress(addr) && srv.ID == raft.ServerID(nodeID) {
				rs.logger.Printf("node %s at %s already member of cluster, ignoring join request", nodeID, addr)
				return nil
			}

			future := rs.RaftServer.RemoveServer(srv.ID, 0, 0)
			if err := future.Error(); err != nil {
				return fmt.Errorf("error removing existing node %s at %s: %s", nodeID, addr, err)
			}
		}
	}

	f := rs.RaftServer.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if f.Error() != nil {
		return f.Error()
	}
	rs.logger.Printf("node %s at %s joined successfully", nodeID, addr)
	return nil
}

func getRaftAddr(raftAddr string) string {
	addrParts := strings.Split(raftAddr, ":")
	httpHost := addrParts[0]
	port, err := strconv.Atoi(addrParts[1])
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s:%d", httpHost, port-1)

}
