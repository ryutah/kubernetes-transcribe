package binding

import (
	"github.com/ryutah/kubernetes-transcribe/pkg/api"
)

// Registry contains the functions needed to support a BindingStorage.
type Registry interface {
	// ApplyBinding should apply the binding. Thas is, it should actually
	// assign or place pod binding.PodID on machine binding.Host.
	ApplyBinding(binding *api.Binding) error
}
