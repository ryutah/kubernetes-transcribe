package apiserver

import (
	"net/http"

	"github.com/ryutah/kubernetes-transcribe/pkg/httplog"
	"github.com/ryutah/kubernetes-transcribe/pkg/runtime"
)

type RedirectHandler struct {
	storage map[string]RESTStorage
	codec   runtime.Codec
}

func (r *RedirectHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	parts := splitPath(req.URL.Path)
	if len(parts) != 2 || req.Method != "GET" {
		notFound(w, req)
		return
	}
	resourceName := parts[0]
	id := parts[1]
	storage, ok := r.storage[resourceName]
	if !ok {
		httplog.LogOf(req, w).Addf("'%v' has no storage object", resourceName)
		notFound(w, req)
		return
	}

	redirector, ok := storage.(Redirector)
	if !ok {
		httplog.LogOf(req, w).Addf("'%v' is not a redirector", resourceName)
		notFound(w, req)
		return
	}

	location, err := redirector.ResourceLocation(id)
	if err != nil {
		status := errToAPIStatus(err)
		writeJSON(status.Code, r.codec, status, w)
		return
	}

	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
