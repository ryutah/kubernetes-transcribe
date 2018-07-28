package health

import (
	"fmt"
	"os/exec"

	"github.com/golang/glog"
	"github.com/ryutah/kubernetes-transcribe/pkg/api"
)

const defaultHealthyRegex = "^OK$"

type CommandRunner interface {
	RunInContainer(podFullName, uuid, containerName string, cmd []string) ([]byte, error)
}

type ExecHealthChecker struct {
	runner CommandRunner
}

func NewExecHealthChecker(runner CommandRunner) HealthChecker {
	return &ExecHealthChecker{runner: runner}
}

func IsExistError(err error) bool {
	_, ok := err.(*exec.ExitError)
	return ok
}

func (e *ExecHealthChecker) HealthCheck(podFullName string, currentState api.PodState, container api.Container) (Status, error) {
	if container.LivenessProbe.Exec == nil {
		return Unknown, fmt.Errorf("Missing exec parameters")
	}
	data, err := e.runner.RunInContainer(podFullName, currentState.Manifest.UUID, container.Name, container.LivenessProbe.Exec.Command)
	glog.V(1).Infof("container %s failed health check: %s", podFullName, string(data))
	if err != nil {
		if IsExistError(err) {
			return Unhealthy, err
		}
		return Unknown, err
	}
	return Healthy, nil
}
