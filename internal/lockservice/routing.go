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
		// fmt.Printf("%s %s\n", rs.httpAddr, string(rs.RaftServer.Leader()))
		// http.Redirect(w, r, string(rs.RaftServer.Leader()), http.StatusOK)
		// return
		url := r.URL
		url.Host, _ = getHTTPAddr(string(rs.RaftServer.Leader()))
		url.Scheme = "http"

		fmt.Println(url.String())

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
	fmt.Printf("Reahed handleAcquire : %v", req)

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
	fmt.Printf("reached handle join %s\n", rs.httpAddr)
	m := map[string]string{}
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	remoteAddr := m["addr"]

	f := rs.RaftServer.AddPeer(raft.ServerAddress(remoteAddr))
	if f.Error() != nil {
		fmt.Println(f.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Printf("no error\n")

	fmt.Printf("node at %s joined successfully", remoteAddr)
	rs.logger.Printf("node at %s joined successfully", remoteAddr)
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
