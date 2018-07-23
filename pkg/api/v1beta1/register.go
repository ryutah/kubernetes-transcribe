package v1beta1

import (
	"github.com/ryutah/kubernetes-transcribe/pkg/api"
	"github.com/ryutah/kubernetes-transcribe/pkg/runtime"
)

// Codec encodes internal objects to the v1beta1 scheme
var Codec = runtime.CodecFor(api.Scheme, "v1beta1")
