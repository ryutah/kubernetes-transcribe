package healthz

import (
	"net/http"
)

type mux interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}

func init() {
	http.HandleFunc("/healthz", handleHealthz)
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	// TODO Support user supplied health functions too.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// InstallHandler registers a handler for health checking on the path "/healthz" to mux.
func InstallHandler(mux mux) {
	mux.HandleFunc("/healthz", handleHealthz)
}
