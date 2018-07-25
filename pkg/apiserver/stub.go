package apiserver

import (
	"net/http"
	"regexp"
)

func CORS(handler http.Handler, alloweedOriginPatterns []*regexp.Regexp, allowedMethods, allowedHeaders []string, allowCredentials string) http.Handler {
	panic("Not implement yet")
}

func RecoverPanics(handler http.Handler) http.Handler {
	panic("Not implement yet")
}
