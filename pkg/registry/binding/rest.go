package binding

import (
	"fmt"

	"github.com/ryutah/kubernetes-transcribe/pkg/api"
	"github.com/ryutah/kubernetes-transcribe/pkg/api/errors"
	"github.com/ryutah/kubernetes-transcribe/pkg/apiserver"
	"github.com/ryutah/kubernetes-transcribe/pkg/labels"
	"github.com/ryutah/kubernetes-transcribe/pkg/runtime"
)

// REST implements the RESTStorage interface for bindings. When bindings are written, it
// changes the location of the affected pods. This information is eventually reflected
// in the pod's CurrentState.Host field.
type REST struct {
	registry Registry
}

// NewREST creates a new REST backed by the given bindingRegistry.
func NewREST(bindingRegistry Registry) *REST {
	return &REST{
		registry: bindingRegistry,
	}
}

// List returns error because bindings are write-only objects.
func (*REST) List(label, field labels.Selector) (runtime.Object, error) {
	return nil, errors.NewNotFound("binding", "list")
}

// Get returns an error because bindings are write-only objects.
func (*REST) Get(id string) (runtime.Object, error) {
	return nil, errors.NewNotFound("binding", id)
}

// Delete returns an error because bindings are write-only objects.
func (*REST) Delete(id string) (<-chan runtime.Object, error) {
	return nil, errors.NewNotFound("binding", id)
}

// New returns a new binding object fit for having data unmarshalled into it.
func (*REST) New() runtime.Object {
	return &api.Binding{}
}

// Create attempts to make the assignment indicated by the binding it receives.
func (b *REST) Create(obj runtime.Object) (<-chan runtime.Object, error) {
	binding, ok := obj.(*api.Binding)
	if !ok {
		return nil, fmt.Errorf("incorrect type: %#v", obj)
	}
	return apiserver.MakeAsync(func() (runtime.Object, error) {
		if err := b.registry.ApplyBinding(binding); err != nil {
			return nil, err
		}
		return &api.Status{Status: api.StatusSuccess}, nil
	}), nil
}

// Update returns an error-- this object may not be updated.
func (*REST) Update(obj runtime.Object) (<-chan runtime.Object, error) {
	return nil, fmt.Errorf("Bindings may not be changed")
}
