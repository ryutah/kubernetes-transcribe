package apiserver

import (
	"net/http"

	"github.com/ryutah/kubernetes-transcribe/pkg/runtime"
)

// ProxyHandler provides a http.Handler which will proxy traffix to locations
// specified by items implementing Redirector.
type ProxyHandler struct {
	prefix  string
	storage map[string]RESTStorage
	codec   runtime.Codec
}

func (p *ProxyHandler) ServeHTTP(http.ResponseWriter, *http.Request) {
	panic("not implemented")
}
