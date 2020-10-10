package routing

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/SystemBuilders/LocKey/internal/lockservice"
)

// acquire wraps the lock Acquire function and creates a clean HTTP service.
func acquire(w http.ResponseWriter, r *http.Request, ls *lockservice.SimpleLockService) {

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
	err = ls.Acquire(desc)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("lock acquired"))
}

func checkAcquired(w http.ResponseWriter, r *http.Request, ls *lockservice.SimpleLockService) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req lockservice.LockCheckRequest
	err = json.Unmarshal(body, &req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	desc := &lockservice.LockDescriptor{
		FileID: req.FileID,
	}

	owner, ok := ls.CheckAcquired(desc)
	if ok {
		byteData, err := json.Marshal(lockservice.CheckAcquireRes{Owner: owner})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Write(byteData)
		return
	}
	http.Error(w, lockservice.ErrCheckAcquireFailure.Error(), http.StatusInternalServerError)
}
