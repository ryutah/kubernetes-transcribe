package client

import (
	"github.com/ryutah/kubernetes-transcribe/pkg/api"
	"github.com/ryutah/kubernetes-transcribe/pkg/labels"
	"github.com/ryutah/kubernetes-transcribe/pkg/util/wait"
)

// ControllerHasDesiredReplicas returns a condition that will be true if the desired replica count
// for a controller's ReplicaSelector equals the Replicas count.
func (c *Client) ControllerHasDesiredReplicas(controller api.ReplicationController) wait.ConditionFunc {
	return func() (bool, error) {
		pods, err := c.ListPods(labels.Set(controller.DesiredState.ReplicaSelector).AsSelector())
		if err != nil {
			return false, err
		}
		return len(pods.Items) == controller.DesiredState.Replicas, nil
	}
}
