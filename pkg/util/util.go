package util

import (
	"encoding/json"
	"fmt"
)

// IntstrKind represents the stored type IntOrString.
type IntstrKind int

const (
	// IntstrInt is the IntOrString holds an int.
	IntstrInt IntstrKind = iota
	// IntstrString is the IntOrString holds a string.
	IntstrString
)

// IntOrString is a type that can hold an int or a string.
// When used in JSON or YAML marshalling and unmarshalling, it produces or consumes the
// inner type. This allows you to have, for example, a JSON field that can accept a name or number.
type IntOrString struct {
	Kind   IntstrKind
	IntVal int
	StrVal string
}

// NewIntOrStringFromInt creates an IntOrString object with an int value.
func NewIntOrStringFromInt(val int) IntOrString {
	return IntOrString{
		Kind:   IntstrInt,
		IntVal: val,
	}
}

// NewIntOrStringFromString creates an IntOrString object with a string value.
func NewIntOrStringFromString(val string) IntOrString {
	return IntOrString{
		Kind:   IntstrString,
		StrVal: val,
	}
}

// SetYAML implements the yaml.Getter interface.
func (intstr *IntOrString) SetYAML(tag string, value interface{}) bool {
	switch v := value.(type) {
	case int:
		intstr.Kind = IntstrInt
		intstr.IntVal = v
		return true
	case string:
		intstr.Kind = IntstrString
		intstr.StrVal = v
		return true
	}
	return false
}

// GetYAML implements the yaml.Getter interface.
func (intstr IntOrString) GetYAML(tag string, value interface{}) {
	switch intstr.Kind {
	case IntstrInt:
		value = intstr.IntVal
	case IntstrString:
		value = intstr.StrVal
	default:
		panic("impossible IntOrString.Kind")
	}
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (intstr *IntOrString) UnmarshalJSON(value []byte) error {
	if value[0] == '"' {
		intstr.Kind = IntstrString
		return json.Unmarshal(value, &intstr.StrVal)
	}
	intstr.Kind = IntstrInt
	return json.Unmarshal(value, &intstr.IntVal)
}

// MarshalJSON implements the json.Marshaller interface.
func (intstr IntOrString) MarshalJSON() ([]byte, error) {
	switch intstr.Kind {
	case IntstrInt:
		return json.Marshal(intstr.IntVal)
	case IntstrString:
		return json.Marshal(intstr.StrVal)
	default:
		return []byte{}, fmt.Errorf("impossible IntOrString.Kind")
	}
}
