package v1beta1

import (
	newer "github.com/ryutah/kubernetes-transcribe/pkg/api"
	"github.com/ryutah/kubernetes-transcribe/pkg/conversion"
	"github.com/ryutah/kubernetes-transcribe/pkg/runtime"
)

func init() {
	newer.Scheme.AddConversionFuncs(
		func(in *newer.EnvVar, out *EnvVar, s conversion.Scope) error {
			out.Value = in.Value
			out.Key = in.Name
			out.Name = in.Name
			return nil
		},
		func(in *EnvVar, out *newer.EnvVar, s conversion.Scope) error {
			out.Value = in.Value
			if in.Name != "" {
				out.Name = in.Name
			} else {
				out.Name = in.Key
			}
			return nil
		},

		// Path & MountType are deprecated.
		func(in *newer.VolumeMount, out *VolumeMount, s conversion.Scope) error {
			out.Name = in.Name
			out.ReadOnly = in.ReadOnly
			out.MountPath = in.MountPath
			out.Path = in.MountPath
			out.MountType = "" // MountType is innored.
			return nil
		},
		func(in *VolumeMount, out *newer.VolumeMount, s conversion.Scope) error {
			out.Name = in.Name
			out.ReadOnly = in.ReadOnly
			if in.MountPath == "" {
				out.MountPath = in.Path
			} else {
				out.MountPath = in.MountPath
			}
			return nil
		},

		// MinionList.Items had a wrong name in v1beta1
		func(in *newer.MinionList, out *MinionList, s conversion.Scope) error {
			s.Convert(&in.JSONBase, &out.JSONBase, 0)
			s.Convert(&in.Items, &out.Items, 0)
			out.Minions = out.Items
			return nil
		},
		func(in *MinionList, out *newer.MinionList, s conversion.Scope) error {
			s.Convert(&in.JSONBase, &out.JSONBase, 0)
			if len(in.Items) == 0 {
				s.Convert(&in.Minions, &out.Items, 0)
			} else {
				s.Convert(&in.Items, &out.Items, 0)
			}
			return nil
		},
	)
}

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
