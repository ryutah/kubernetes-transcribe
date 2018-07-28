package apiserver

import (
	"fmt"
	"net/http"

	"github.com/ryutah/kubernetes-transcribe/pkg/api"
	"github.com/ryutah/kubernetes-transcribe/pkg/tools"
)

// statusError is an object that can be converted into an api.Status
type statusError interface {
	Status() api.Status
}

// errToAPIStatus converts an error to an api.Status object.
func errToAPIStatus(err error) *api.Status {
	switch t := err.(type) {
	case statusError:
		status := t.Status()
		status.Status = api.StatusFailure
		// TODO: check for invalid responses
		return &status
	default:
		status := http.StatusInternalServerError
		switch {
		// TODO: replace me with NewConflictErr
		case tools.IsEtcdTestFailed(err):
			status = http.StatusConflict
		}
		return &api.Status{
			Status:  api.StatusFailure,
			Code:    status,
			Reason:  api.StatusReasonUnknown,
			Message: err.Error(),
		}
	}
}

// notFound renders a simple not found error.
func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "Not found: %#v", r.RequestURI)
}

// badGatewayError renders a simple bad gateway error.
func badGatewayError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprintf(w, "Bad Gateway: %#v", r.RequestURI)
}
