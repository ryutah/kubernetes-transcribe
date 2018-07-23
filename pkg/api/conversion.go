package api

import (
	"github.com/ryutah/kubernetes-transcribe/pkg/runtime"
)

// Codec is the identity codec for this package - it can only convert itself
// to itself.
var Codec = runtime.CodecFor(Scheme, "")

// EmbeddedObject implements a Codec specific version of an
// embedded object.
type EmbeddedObject struct {
	runtime.Object
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (a *EmbeddedObject) UnmarshalJSON(b []byte) error {
	obj, err := runtime.CodecUnmarshalJSON(Codec, b)
	a.Object = obj
	return err
}

// MarshalJSON implements the json.Marshaler interface.
func (a EmbeddedObject) MarshalJSON() ([]byte, error) {
	return runtime.CodecMarshalJSON(Codec, a.Object)
}

// SetYAML implements the yaml.Setter interface.
func (a *EmbeddedObject) SetYAML(tag string, value interface{}) bool {
	obj, ok := runtime.CodecSetYAML(Codec, tag, value)
	a.Object = obj
	return ok
}

// GetYAML implements the yaml.Getter interface.
func (a EmbeddedObject) GetYAML() (tag string, value interface{}) {
	return runtime.CodecGetYAML(Codec, a.Object)
}
