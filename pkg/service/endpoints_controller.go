package service

import (
	"fmt"
	"net"
	"strconv"

	"github.com/golang/glog"
	"github.com/ryutah/kubernetes-transcribe/pkg/api"
	"github.com/ryutah/kubernetes-transcribe/pkg/client"
	"github.com/ryutah/kubernetes-transcribe/pkg/labels"
	"github.com/ryutah/kubernetes-transcribe/pkg/registry/service"
	"github.com/ryutah/kubernetes-transcribe/pkg/util"
)

// EndpointController manages service endpoints.
type EndpointController struct {
	client          *client.Client
	serviceRegistry service.Registry
}

// NewEndpointController returns a new *EndpointController
func NewEndpointController(serviceRegistry service.Registry, client *client.Client) *EndpointController {
	return &EndpointController{
		serviceRegistry: serviceRegistry,
		client:          client,
	}
}

// SyncServiceEndpoints syncs service endpoints.
func (e *EndpointController) SyncServiceEndpoints() error {
	services, err := e.client.ListServices(labels.Everything())
	if err != nil {
		glog.Errorf("Failed to list services: %v", err)
		return err
	}
	var resultErr error
	for _, service := range services.Items {
		pods, err := e.client.ListPods(labels.Set(service.Selector).AsSelector())
		if err != nil {
			glog.Errorf("Error syncing service: %#v, skipping.", service)
			resultErr = err
			continue
		}
		endpoints := make([]string, len(pods.Items))
		for ix, pod := range pods.Items {
			port, err := findPort(&pod.DesiredState.Manifest, service.ContainerPort)
			if err != nil {
				glog.Errorf("Failed to find port for service: %v, %v", service, err)
				continue
			}
			if len(pod.CurrentState.PodIP) == 0 {
				glog.Errorf("Failed to find an IP for pod: %v", pod)
				continue
			}
			endpoints[ix] = net.JoinHostPort(pod.CurrentState.PodIP, strconv.Itoa(port))
		}
		err = e.serviceRegistry.UpdateEndpoints(&api.Endpoints{
			JSONBase:  api.JSONBase{ID: service.ID},
			Endpoints: endpoints,
		})
		// TODO: this is totally broken, we need to compute this and store inside an AtomicUpdate loop.
		if err != nil {
			glog.Errorf("Error updating endpoints: %#v", err)
			continue
		}
	}
	return resultErr
}

// findPort locates the container port for the given manifest and portName.
func findPort(manifest *api.ContainerManifest, portName util.IntOrString) (int, error) {
	if ((portName.Kind == util.IntstrString && len(portName.StrVal) == 0) ||
		(portName.Kind == util.IntstrInt && portName.IntVal == 0)) &&
		len(manifest.Containers[0].Ports) > 0 {
		return manifest.Containers[0].Ports[0].ContainerPort, nil
	}
	if portName.Kind == util.IntstrInt {
		return portName.IntVal, nil
	}
	name := portName.StrVal
	for _, container := range manifest.Containers {
		for _, port := range container.Ports {
			if port.Name == name {
				return port.ContainerPort, nil
			}
		}
	}
	return -1, fmt.Errorf("no suitable port for manifest: %s", manifest.ID)
}
