package apiserver

import (
	"fmt"
	"net/http"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/golang/glog"
	"github.com/ryutah/kubernetes-transcribe/pkg/httplog"
)

// RecoverPanics wraps http.Handler to recover and log panics.
func RecoverPanics(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if x := recover(); x != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, "apis panic. Look in log for details.")
				glog.Infof("APIServer panic'd on %v %v: %#v\n%s\n", req.Method, req.RequestURI, x, debug.Stack())
			}
		}()
		defer httplog.NewLogged(req, &w).StacktraceWhen(
			httplog.StatusIsNot(
				http.StatusOK,
				http.StatusAccepted,
				http.StatusMovedPermanently,
				http.StatusTemporaryRedirect,
				http.StatusConflict,
				http.StatusNotFound,
				StatusUnprocessableEntity,
			),
		).Log()

		// Dispatch to the internal handler
		handler.ServeHTTP(w, req)
	})
}

// Simple CORS implementation that wraps an http Handler
// For a more detailed implementation use https://github.com/martini-contrib/cors
// or implement CORS at your proxy layer
// Pass nil for allowedMethods and allowedHeaders to use the defaults
func CORS(handler http.Handler, alloweedOriginPatterns []*regexp.Regexp, allowedMethods, allowedHeaders []string, allowCredentials string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		origin := req.Header.Get("Origin")
		if origin != "" {
			allowed := false
			for _, pattern := range alloweedOriginPatterns {
				if allowed = pattern.MatchString(origin); allowed {
					break
				}
			}
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				// Set defaults for methods and headers if nothing was passed
				if allowedMethods == nil {
					allowedMethods = []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"}
				}
				if allowedHeaders == nil {
					allowedHeaders = []string{"Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "X-Requested-With", "If-Modified-Since"}
				}
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
				w.Header().Set("Access-Control-Allow-Credentials", allowCredentials)

				// Stop here if its a preflight OPTIONS request
				if req.Method == "OPTIONS" {
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
		}
		// Dispatch to the next handler
		handler.ServeHTTP(w, req)
	})
}
