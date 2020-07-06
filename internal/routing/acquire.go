package routing

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/GoPlayAndFun/LocKey/internal/lockservice"
)

// type Request = lockservice.Request

func acquire(w http.ResponseWriter, r *http.Request, ls *lockservice.SimpleLockService) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req lockservice.Request
	err = json.Unmarshal(body, &req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	desc := &lockservice.SimpleDescriptor{
		FileID: req.FileID,
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

	desc := &lockservice.SimpleDescriptor{
		FileID: string(body),
	}
	if ls.CheckAcquired(desc) {
		w.Write([]byte("checkAcquire success"))
		return
	}
	w.Write([]byte("checkAcquire failure"))
}
