package routing

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/SystemBuilders/LocKey/internal/lockservice"
)

func release(w http.ResponseWriter, r *http.Request, ls *lockservice.SimpleLockService) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req lockservice.LockRequest
	err = json.Unmarshal(body, &req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	desc := &lockservice.LockDescriptor{
		FileID: req.FileID,
		UserID: req.UserID,
	}
	err = ls.Release(desc)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("lock released"))
}

func checkReleased(w http.ResponseWriter, r *http.Request, ls *lockservice.SimpleLockService) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req lockservice.LockRequest
	err = json.Unmarshal(body, &req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	desc := &lockservice.LockDescriptor{
		FileID: req.FileID,
		UserID: req.UserID,
	}

	if ls.CheckReleased(desc) {
		w.Write([]byte("checkRelease success"))
		return
	}
	w.Write([]byte("checkRelease failure"))

}
