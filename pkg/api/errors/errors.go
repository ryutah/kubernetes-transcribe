package errors

import (
	"fmt"
	"net/http"

	"github.com/ryutah/kubernetes-transcribe/pkg/api"
)

// statusError is an error intended for consumption by a REST API server.
type statusError struct {
	status api.Status
}

func (s *statusError) Error() string {
	return s.status.Message
}

func (s *statusError) Status() api.Status {
	return s.status
}

// NewNotFound returns a new error which indicates that the resource of the kind and the name was not found.
func NewNotFound(kind, name string) error {
	return &statusError{
		api.Status{
			Status: api.StatusFailure,
			Code:   http.StatusNotFound,
			Reason: api.StatusReasonNotFound,
			Details: &api.StatusDetails{
				Kind: kind,
				ID:   name,
			},
			Message: fmt.Sprintf("%s %q not found", kind, name),
		},
	}
}

// NewAlreadyExists returns an error indicating the item requested exists by that identifier.
func NewAlreadyExists(kind, name string) error {
	return &statusError{
		api.Status{
			Status: api.StatusFailure,
			Code:   http.StatusConflict,
			Reason: api.StatusReasonAlreadyExists,
			Details: &api.StatusDetails{
				Kind: kind,
				ID:   name,
			},
			Message: fmt.Sprintf("%s %q already exists", kind, name),
		},
	}
}

// NewConflict returns an error indicating the item can't be updated as provided.
func NewConflict(kind, name string, err error) error {
	return &statusError{
		api.Status{
			Status: api.StatusFailure,
			Code:   http.StatusConflict,
			Reason: api.StatusReasonConflict,
			Details: &api.StatusDetails{
				Kind: kind,
				ID:   name,
			},
			Message: fmt.Sprintf("%s %q cannot be updated: %s", kind, name, err),
		},
	}
}

// NewInvalid returns an error indicating the item is invalid and cannnot be processed.
func NewInvalid(kind, name string, errs ErrorList) error {
	causes := make([]api.StatusCause, 0, len(errs))
	for _, err := range errs {
		if err, ok := err.(ValidationError); ok {
			causes = append(causes, api.StatusCause{
				Type:    api.CauseType(err.Type),
				Message: err.Error(),
				Field:   err.Field,
			})
		}
	}
	return &statusError{
		api.Status{
			Status: api.StatusFailure,
			Code:   http.StatusUnprocessableEntity,
			Reason: api.StatusReasonInvalid,
			Details: &api.StatusDetails{
				Kind:   kind,
				ID:     name,
				Causes: causes,
			},
			Message: fmt.Sprintf("%s %q is invalid: %s", kind, name, errs.ToError()),
		},
	}
}

// IsNotFound returns true if the specified error was created by NewNotFoundErr.
func IsNotFound(err error) bool {
	return reasonForError(err) == api.StatusReasonNotFound
}

// IsAlreadyExists determines if the err is an error which indicates that a specified resource already exists.
func IsAlreadyExists(err error) bool {
	return reasonForError(err) == api.StatusReasonAlreadyExists
}

// IsConflict determines if the err is an error which indicates the provided update conflicts.
func IsConflict(err error) bool {
	return reasonForError(err) == api.StatusReasonConflict
}

// IsInvalid determines if the err is an error which indicates the provided resource is not valid.
func IsInvalid(err error) bool {
	return reasonForError(err) == api.StatusReasonInvalid
}

func reasonForError(err error) api.StatusReason {
	switch t := err.(type) {
	case *statusError:
		return t.status.Reason
	}
	return api.StatusReasonUnknown
}
