package routing

import (
	"net/http"

	"github.com/GoPlayAndFun/Distributed-File-System/internal/lockservice"
	"github.com/gorilla/mux"
)

// SetupRouting adds all the routes on the http server.
func SetupRouting(ls *lockservice.SimpleLockService, r *mux.Router) *mux.Router {
	r.HandleFunc("/acquire", makeAcquireHandler(ls)).Methods(http.MethodPost)
	r.HandleFunc("/checkAcquire", makeCheckAcquiredHandler(ls)).Methods(http.MethodPost)
	r.HandleFunc("/release", makeReleaseHandler(ls)).Methods(http.MethodPost)
	r.HandleFunc("/checkRelease", makeCheckReleaseHandler(ls)).Methods(http.MethodPost)
	return r
}

func makeAcquireHandler(ls *lockservice.SimpleLockService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		acquire(w, r, ls)
	}
}

func makeCheckAcquiredHandler(ls *lockservice.SimpleLockService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		checkAcquired(w, r, ls)
	}
}

func makeReleaseHandler(ls *lockservice.SimpleLockService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		release(w, r, ls)
	}
}

func makeCheckReleaseHandler(ls *lockservice.SimpleLockService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		checkReleased(w, r, ls)
	}
}
