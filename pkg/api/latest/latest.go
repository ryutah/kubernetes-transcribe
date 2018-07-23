package latest

import (
	"github.com/ryutah/kubernetes-transcribe/pkg/api/v1beta1"
	"github.com/ryutah/kubernetes-transcribe/pkg/runtime"
)

// Version is the string that represents the current external default version.
const Version = "v1beta1"

// Versions is the list of verions that are recognized in code. The order provided
// may be assumed to be latest feature rich to most feature righ, and clients may
// choose to prefer the latter iterms in the list over the former items when presented
// with a set of versions to choose.
var Versions = []string{"v1beta1", "v1beta2"}

// Codec is the default codec for serializing output that should use
// the latest supported version. Use this Codec when writing to
// disk, a data store that is not dynamically versioned, or in tests.
// This codec can decode any object that Kubernrtes is aware of.
var Codec = v1beta1.Codec

// TODO: Not implement yet.
func InterfacesFor(version string) (codec runtime.Codec, versioner runtime.ResourceVersioner, err error) {
	return nil, nil, nil
}
