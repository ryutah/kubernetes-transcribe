package client

import (
	"github.com/ryutah/kubernetes-transcribe/pkg/api"
	"github.com/ryutah/kubernetes-transcribe/pkg/labels"
)

type Client struct{}

// TODO: Not implements yet.
func (c *Client) ListServices(selector labels.Selector) (result *api.ServiceList, err error) {
	return nil, nil
}

// TODO: Not implements yet.
func (c *Client) ListPods(selector labels.Selector) (result *api.PodList, err error) {
	return nil, nil
}
