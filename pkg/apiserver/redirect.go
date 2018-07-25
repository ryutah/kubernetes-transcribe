package apiserver

import (
	"net/http"

	"github.com/ryutah/kubernetes-transcribe/pkg/runtime"
)

type RedirectHandler struct {
	storage map[string]RESTStorage
	codec   runtime.Codec
}

func (r *RedirectHandler) ServeHTTP(http.ResponseWriter, *http.Request) {
	panic("not implemented")
}
