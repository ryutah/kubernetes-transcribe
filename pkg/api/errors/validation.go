package errors

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
)

// ValidationErrorType is a machine readable value providing more detail about why
// a field is invalid.  These values are expected to match 1-1 with
// CauseType in api/tpyes.go
type ValidationErrorType string

const (
	// ValidationErrorTypeNotFound is used to report failure to find a requested value
	// (e.g. looking up an ID).
	ValidationErrorTypeNotFound ValidationErrorType = "fieldValueNotFound"
	// ValidationErrorTypeRequired is used to report required values that are not
	// provided (e.g. empty strings, null values, or empty arrays).
	ValidationErrorTypeRequired ValidationErrorType = "fieldValueRequired"
	// ValidationErrorTypeDuplicate is used to report collisions of values that must be
	// unique (e.g. unique IDs).
	ValidationErrorTypeDuplicate ValidationErrorType = "fieldValueDuplicate"
	// ValidationErrorTypeInvalid is used to report malformed values (e.g. failed regex
	// match).
	ValidationErrorTypeInvalid ValidationErrorType = "fieldValueInvalid"
	// ValidationErrorTypeNotSupported is used to report valid (as per formatting rules)
	// values that can not be handled (e.g. an enumerated string).
	ValidationErrorTypeNotSupported ValidationErrorType = "fieldValueNotSupported"
)

func ValueOf(t ValidationErrorType) string {
	switch t {
	case ValidationErrorTypeNotFound:
		return "not found"
	case ValidationErrorTypeRequired:
		return "required value"
	case ValidationErrorTypeDuplicate:
		return "duplicate value"
	case ValidationErrorTypeInvalid:
		return "invalid value"
	case ValidationErrorTypeNotSupported:
		return "unsupported value"
	default:
		glog.Errorf("unrecognized validation type: %#v", t)
		return ""
	}
}

// ValidationError is an implementation of the 'error' interface, which represents an error of validation.
type ValidationError struct {
	Type     ValidationErrorType
	Field    string
	BadValue interface{}
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("%s: %v '%v'", v.Field, ValueOf(v.Type), v.BadValue)
}

// NewFieldRequired returns a ValidationError indicating "value required"
func NewFieldRequired(field string, value interface{}) ValidationError {
	return ValidationError{
		Type:     ValidationErrorTypeRequired,
		Field:    field,
		BadValue: value,
	}
}

// NewFieldInvalid returns a ValidationError indicating "invalid value"
func NewFieldInvalid(field string, value interface{}) ValidationError {
	return ValidationError{
		Type:     ValidationErrorTypeInvalid,
		Field:    field,
		BadValue: value,
	}
}

// NewFieldNotSupported returns a ValidationError indicating "unsupported value"
func NewFieldNotSupported(field string, value interface{}) ValidationError {
	return ValidationError{
		Type:     ValidationErrorTypeNotSupported,
		Field:    field,
		BadValue: value,
	}
}

// NewFieldDuplicate returns a ValidationError indicating "duplicate value"
func NewFieldDuplicate(field string, value interface{}) ValidationError {
	return ValidationError{
		Type:     ValidationErrorTypeDuplicate,
		Field:    field,
		BadValue: value,
	}
}

// NewFieldNotFound returns a ValidationError indicating "value not found"
func NewFieldNotFound(field string, value interface{}) ValidationError {
	return ValidationError{
		Type:     ValidationErrorTypeNotFound,
		Field:    field,
		BadValue: value,
	}
}

// ErrorList is a collection of errors.  This does not implement the error
// interface to avoid confusing where an empty ErrorList would still be an
// error (non-nil).  To produce a single error instance from an ErrorList, use
// the ToError() method, which will return nil for an empty ErrorList.
type ErrorList []error

// This helper implements the error interface for ErrorList, but prevents
// accidental conversion of ErrorList to error.
type errorListInternal ErrorList

func (list errorListInternal) Error() string {
	if len(list) == 0 {
		return ""
	}
	sl := make([]string, len(list))
	for i := range list {
		sl[i] = list[i].Error()
	}
	return strings.Join(sl, "; ")
}

// ToError converts an ErrorList a "normal" error, or nil if the list is empty.
func (list ErrorList) ToError() error {
	if len(list) == 0 {
		return nil
	}
	return errorListInternal(list)
}

// Prefix adds a prefix to the Field of every ValidationError in the list. Returns
// the list for convenience.
func (list ErrorList) Prefix(prefix string) ErrorList {
	for i := range list {
		if err, ok := list[i].(ValidationError); ok {
			if strings.HasPrefix(err.Field, "[") {
				err.Field = prefix + err.Field
			} else if len(err.Field) != 0 {
				err.Field = prefix + "." + err.Field
			} else {
				err.Field = prefix
			}
			list[i] = err
		}
	}
	return list
}

// PrefixIndex adds an index to the Field of every ValidationError in the list. Returns
// the list for convenience.
func (list ErrorList) PrefixIndex(index int) ErrorList {
	return list.Prefix(fmt.Sprintf("[%d]", index))
}
