package runtime

import (
	"github.com/ryutah/kubernetes-transcribe/pkg/util"
)

// Note that the types provided in this file are not versioned and are intended to be
// safe to use from within all versions of every API object.

// JSONBase is shared by all top level objects. The proper way to use it is to inline it in your type.
// link this:
// type MyAwesomeAPIObject struct {
//	runtime.JSONBase `yaml:",inline" json:",inline"`
//	... // other fields
// }
// func (*MyAwesomeAPIObject) IsAnAPIObject() {}
//
// JSONBase is provided here for convenience. You may use it directly from this package or define
// your own with the same fields.
type JSONBase struct {
	Kind              string    `json:"kind,omitempty" yaml:"kind,omitempty"`
	ID                string    `json:"id,omitempty" yaml:"id,omitempty"`
	CreationTimestamp util.Time `json:"creationTimestamp,omitempty" yaml:"creationTimestamp,omitempty"`
	SelfLink          string    `json:"selfLink,omitempty" yaml:"selfLink,omitempty"`
	ResourceVersion   uint64    `json:"resourceVersion,omitempty" yaml:"resourceVersion,omitempty"`
	APIVersion        string    `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
}

// PluginBase is like JSONBase, but it's intended for plugin objects that won't ever be encoded
// except while embedded in other objects.
type PluginBase struct {
	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`
}

// EmbeddedObject has appropriate encoder and decoder functions, such that on the wire, it's
// stored as a []byte, but in memory, the containerd object is accessable as an Object
// via the Get() function. Only valid API objects may be stored via EmbeddedObject.
// The purpose of this is to allow an API object of type known only at runtime to be
// embedded within other API objects.
//
// Note that object assumes that you've registered all of your api types with the api package.
// EmbeddedObject and RawExtension can be used together to allow for API object extensions:
// see the comment for RawExtension.
type EmbeddedObject struct {
	Object
}

// RawExtension is used with EmbeddedObject to do a two-phase encoding of extension objects.
// To use this, make a field which has RawExtension as its type in your external, versioned
// struct, and EmbeddedObject in your internal struct. You also need to register your various plugin types.
//
// // Internal pacakge:
// type MyAPIObject struct {
//	runtime.JSONBase `yaml:",inline" json:",inline"`
//	MyPlugin runtime.EmbeddedObject `json:"myPlugin" yaml:"myPlugin"`
// }
// type PluginA struct {
//	runtime.PluginBase `yaml:",inline" json:",inline"`
//	AOption string `yaml:"aOption" json:"aOption"`
// }
//
// // External package:
// type MyAPIObject struct {
//	runtime.JSONBase `yaml:",inline" json:",inline"`
//	MyPlugin runtime.RawExtension `json:"myPlugin" yaml:"myPlugin"`
// }
// type PluginA struct {
//	runtime.PluginBase `yaml:",inline" json:",inline"`
//	AOption string `yaml:"aOption" json:"aOption"`
// }
//
// // On the wire, the JSON will look something like this:
// {
//   "kind": "MyAPIObject",
//   "apiVersion": "v1beta1",
//   "myPlugin": {
//     "kind": "PluginA",
//     "aOption": "foo",
//   }
// }
//
// So what happens? Decode first uses json or yaml to unmarshal the serialized data into your external MyAPIObject. That causes the raw JSON to be stored, but not unpackaged.
// The next step is to copy (using pkg/conversion) into the internal struct. The runtime
// package's DefaultScheme has conversion functions installed which will unpack the
// JSON stored in RawExtension, turning it into the correct object type, and storing it
// in the EmbeddedObject. (TODO: in the case where the object is of an unknown type, a
// runtime.Unknown object will be created and stored.)
type RawExtension struct {
	RawJSON []byte
}

// Unknown allows api objects with unknown types to be passed-through. This can be used
// to deal with the API objects from a plug-in. Unknown objects still have functioning
// JSONBase features-- kind, version, resourceVersion, etc.
// TODO: Not implemented yet
type Unknown struct {
	JSONBase `yaml:",inline" json:",inline"`
	// RawJSON will hold the complete JSON of the object which couldn't be matched
	// with a registered type. Most likely, nothing should be done with this
	// except for passing it through the system.
	RawJSON []byte
}

func (*Unknown) IsAnAPIObject() {}
