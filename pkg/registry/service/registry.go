package service

import (
	"github.com/ryutah/kubernetes-transcribe/pkg/api"
	"github.com/ryutah/kubernetes-transcribe/pkg/labels"
	"github.com/ryutah/kubernetes-transcribe/pkg/registry/endpoint"
	"github.com/ryutah/kubernetes-transcribe/pkg/watch"
)

// Registry is an interface for things that konw how to store services.
type Registry interface {
	ListServices() (*api.ServiceList, error)
	CreateService(svc *api.Service) error
	GetService(name string) (*api.Service, error)
	DeleteService(name string) error
	UpdateService(svc *api.Service) error
	WatchServices(labels, fields labels.Selector, resourceVersion uint64) (watch.Interface, error)

	// TODO: endpoints and their implementation should be separated, setting endpoints should be
	// supported via the API, and the endpoints-controller should use the API to update endpoints.
	endpoint.Registry
}
