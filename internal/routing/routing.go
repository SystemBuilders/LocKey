package routing

import (
	"net/http"

	"github.com/SystemBuilders/LocKey/internal/lockservice"
	"github.com/gorilla/mux"
)

// SetupRouting adds all the routes on the http server.
func SetupRouting(ls *lockservice.SimpleLockService, r *mux.Router) *mux.Router {
	r.HandleFunc("/acquire", makeacquireHandler(ls)).Methods(http.MethodPost)
	r.HandleFunc("/checkAcquire", makecheckAcquiredHandler(ls)).Methods(http.MethodPost)
	r.HandleFunc("/release", makereleaseHandler(ls)).Methods(http.MethodPost)
	r.HandleFunc("/checkRelease", makecheckReleaseHandler(ls)).Methods(http.MethodPost)
	return r
}

func makeacquireHandler(ls *lockservice.SimpleLockService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		acquire(w, r, ls)
	}
}

func makecheckAcquiredHandler(ls *lockservice.SimpleLockService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		checkAcquired(w, r, ls)
	}
}

func makereleaseHandler(ls *lockservice.SimpleLockService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		release(w, r, ls)
	}
}

func makecheckReleaseHandler(ls *lockservice.SimpleLockService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		checkReleased(w, r, ls)
	}
}
