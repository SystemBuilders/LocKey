package routing

import (
	"io/ioutil"
	"net/http"

	"github.com/GoPlayAndFun/Distributed-File-System/internal/lockservice"
)

func acquire(w http.ResponseWriter, r *http.Request, ls *lockservice.SimpleLockService) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	desc := &lockservice.SimpleDescriptor{
		FileID: string(body),
	}

	err = ls.Acquire(desc)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
