package routing

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/GoPlayAndFun/LocKey/internal/lockservice"
)

func acquire(w http.ResponseWriter, r *http.Request, ls *lockservice.SimpleLockService) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var req interface{}
	err = json.Unmarshal(body, &req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := req.(map[string]interface{})

	for k, v := range data {
		switch v := v.(type) {
		case string:
			if k == "FileID" {
				desc := &lockservice.SimpleDescriptor{
					FileID: string(v),
				}
				err = ls.Acquire(desc)
			}
		}
	}

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
