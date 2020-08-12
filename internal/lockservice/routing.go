package lockservice

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

func (rs *RaftStore) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "/acquire") {
		rs.handleAcquire(w, r)
	} else if strings.Contains(r.URL.Path, "/release") {
		rs.handleRelease(w, r)
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
