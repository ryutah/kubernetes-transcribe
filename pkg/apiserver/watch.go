package apiserver

import (
	"net/http"

	"github.com/ryutah/kubernetes-transcribe/pkg/runtime"
)

type WatchHandler struct {
	storage map[string]RESTStorage
	codec   runtime.Codec
}

func (w *WatchHandler) ServeHTTP(http.ResponseWriter, *http.Request) {
	panic("not implemented")
}
