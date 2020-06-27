package routing

import (
	"io/ioutil"
	"net/http"

	"github.com/GoPlayAndFun/Distributed-File-System/internal/lockservice"
)

func release(w http.ResponseWriter, r *http.Request, ls *lockservice.SimpleLockService) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	desc := &lockservice.SimpleDescriptor{
		FileID: string(body),
	}

	err = ls.Release(desc)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func checkReleased(w http.ResponseWriter, r *http.Request, ls *lockservice.SimpleLockService) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	desc := &lockservice.SimpleDescriptor{
		FileID: string(body),
	}

	if ls.CheckReleased(desc) {
		w.Write([]byte("checkRelease success"))
		return
	}
	w.Write([]byte("checkRelease failure"))

}
