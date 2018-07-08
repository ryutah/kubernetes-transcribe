package apiserver

import (
	"net/http"
	"regexp"

	"github.com/ryutah/kubernetes-transcribe/pkg/runtime"
)

type RESTStorage interface{}

type APIGroup struct{}

func (g *APIGroup) InstallREST(m mux, paths ...string) {
	panic("Not implement yet")
}

type mux interface{}

func NewAPIGroup(storage map[string]RESTStorage, codec runtime.Codec) *APIGroup {
	panic("Not implement yet")
}

func InstallSupport(m mux) {
	panic("Not implement yet")
}

func CORS(handler http.Handler, alloweedOriginPatterns []*regexp.Regexp, allowedMethods, allowedHeaders []string, allowCredentials string) http.Handler {
	panic("Not implement yet")
}

func RecoverPanics(handler http.Handler) http.Handler {
	panic("Not implement yet")
}
