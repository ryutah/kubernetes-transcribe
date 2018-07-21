package endpoint

import (
	"github.com/ryutah/kubernetes-transcribe/pkg/api"
	"github.com/ryutah/kubernetes-transcribe/pkg/labels"
	"github.com/ryutah/kubernetes-transcribe/pkg/watch"
)

// Registry is an interface for things that know how to store endpoints.
type Registry interface {
	ListEndpoints() (*api.EndpointsList, error)
	GetEndpoints(name string) (*api.Endpoints, error)
	WatchEndpoints(labels, fields labels.Selector, resourceVersion uint64) (watch.Interface, error)
	UpdateEndpoints(s *api.Endpoints) error
}
