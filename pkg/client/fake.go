package client

import (
	"github.com/ryutah/kubernetes-transcribe/pkg/api"
	"github.com/ryutah/kubernetes-transcribe/pkg/labels"
	"github.com/ryutah/kubernetes-transcribe/pkg/version"
	"github.com/ryutah/kubernetes-transcribe/pkg/watch"
)

type FakeAction struct {
	Action string
	Value  interface{}
}

// Fake implements Interface. Meant to be embedded into a struct to get a default
// implementation. This makes faking out just the method you want to test easier.
type Fake struct {
	// Fake by default keeps a simple list of the methods that have been called.
	Actions       []FakeAction
	Pods          api.PodList
	Ctrl          api.ReplicationController
	ServiceList   api.ServiceList
	EndpointsList api.EndpointsList
	Minions       api.MinionList
	Err           error
	Watch         watch.Interface
}

func (c *Fake) ListPods(selector labels.Selector) (*api.PodList, error) {
	c.Actions = append(c.Actions, FakeAction{Action: "list-pods"})
	return api.Scheme.CopyOrDie(&c.Pods).(*api.PodList), nil
}

func (c *Fake) GetPod(name string) (*api.Pod, error) {
	c.Actions = append(c.Actions, FakeAction{Action: "get-pod", Value: name})
	return &api.Pod{}, nil
}

func (c *Fake) DeletePod(name string) error {
	c.Actions = append(c.Actions, FakeAction{Action: "delete-pod", Value: name})
	return nil
}

func (c *Fake) CreatePod(*api.Pod) (*api.Pod, error) {
	c.Actions = append(c.Actions, FakeAction{Action: "create-pod"})
	return &api.Pod{}, nil
}

func (c *Fake) UpdatePod(pod *api.Pod) (*api.Pod, error) {
	c.Actions = append(c.Actions, FakeAction{Action: "update-pod", Value: pod.ID})
	return &api.Pod{}, nil
}

func (c *Fake) ListReplicationControllers(selector labels.Selector) (*api.ReplicationControllerList, error) {
	c.Actions = append(c.Actions, FakeAction{Action: "list-controllers"})
	return &api.ReplicationControllerList{}, nil
}

func (c *Fake) GetReplicationController(name string) (*api.ReplicationController, error) {
	c.Actions = append(c.Actions, FakeAction{Action: "get-controller", Value: name})
	return api.Scheme.CopyOrDie(&c.Ctrl).(*api.ReplicationController), nil
}

func (c *Fake) CreateReplicationController(controller *api.ReplicationController) (*api.ReplicationController, error) {
	c.Actions = append(c.Actions, FakeAction{Action: "create-controller", Value: controller})
	return &api.ReplicationController{}, nil
}

func (c *Fake) UpdateReplicationController(controller *api.ReplicationController) (*api.ReplicationController, error) {
	c.Actions = append(c.Actions, FakeAction{Action: "update-controller", Value: controller})
	return &api.ReplicationController{}, nil
}

func (c *Fake) DeleteReplicationController(controller string) error {
	c.Actions = append(c.Actions, FakeAction{Action: "delete-controller", Value: controller})
	return nil
}

func (c *Fake) WatchReplicationControllers(label labels.Selector, field labels.Selector, resourceVersion uint64) (watch.Interface, error) {
	c.Actions = append(c.Actions, FakeAction{Action: "watch-controllers", Value: resourceVersion})
	return c.Watch, nil
}

func (c *Fake) ListServices(selector labels.Selector) (*api.ServiceList, error) {
	c.Actions = append(c.Actions, FakeAction{Action: "list-services"})
	return &c.ServiceList, c.Err
}

func (c *Fake) GetService(name string) (*api.Service, error) {
	c.Actions = append(c.Actions, FakeAction{Action: "get-service", Value: name})
	return &api.Service{}, nil
}

func (c *Fake) CreateService(service *api.Service) (*api.Service, error) {
	c.Actions = append(c.Actions, FakeAction{Action: "create-service", Value: service})
	return &api.Service{}, nil
}

func (c *Fake) UpdateService(service *api.Service) (*api.Service, error) {
	c.Actions = append(c.Actions, FakeAction{Action: "update-service", Value: service})
	return &api.Service{}, nil
}

func (c *Fake) DeleteService(service string) error {
	c.Actions = append(c.Actions, FakeAction{Action: "delete-service", Value: service})
	return nil
}

func (c *Fake) WatchServices(label labels.Selector, field labels.Selector, resourceVersion uint64) (watch.Interface, error) {
	c.Actions = append(c.Actions, FakeAction{Action: "watch-services", Value: resourceVersion})
	return c.Watch, c.Err
}

func (c *Fake) ListEndpoints(selector labels.Selector) (*api.EndpointsList, error) {
	c.Actions = append(c.Actions, FakeAction{Action: "list-endpoints"})
	return api.Scheme.CopyOrDie(&c.EndpointsList).(*api.EndpointsList), c.Err
}

func (c *Fake) WatchEndpoints(label, field labels.Selector, resourceVersion uint64) (watch.Interface, error) {
	c.Actions = append(c.Actions, FakeAction{Action: "watch-endpoints", Value: resourceVersion})
	return c.Watch, c.Err
}

func (c *Fake) ServiceVersion() (*version.Info, error) {
	c.Actions = append(c.Actions, FakeAction{Action: "get-version", Value: nil})
	versionInfo := version.Get()
	return &versionInfo, nil
}

func (c *Fake) ListMinions() (*api.MinionList, error) {
	c.Actions = append(c.Actions, FakeAction{Action: "list-minions", Value: nil})
	return &c.Minions, nil
}
