package apiserver

import (
	"net/http"
	"time"

	"github.com/ryutah/kubernetes-transcribe/pkg/runtime"
)

type RESTHandler struct {
	storage     map[string]RESTStorage
	codec       runtime.Codec
	ops         *Operations
	asyncOpWait time.Duration
}

func (r *RESTHandler) ServeHTTP(http.ResponseWriter, *http.Request) {
	panic("not implemented")
}
