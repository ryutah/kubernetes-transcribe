package health

import (
	"sync"

	"github.com/golang/glog"
	"github.com/ryutah/kubernetes-transcribe/pkg/api"
)

// Status represents the result of a single health-check operation.
type Status int

// Status values must be one of these constants.
const (
	Healthy Status = iota
	Unhealthy
	Unknown
)

// HealthChecker defines an abstract interface for checking container health.
type HealthChecker interface {
	HealthCheck(podFullName string, currentState api.PodState, container api.Container) (Status, error)
}

// protects checkers
var checkerLock = sync.Mutex{}
var checkers = map[string]HealthChecker{}

// AddHealthChecker adds a health checker to the list of known HealthChecker objects.
// Any subsequent call to NewHealthChecker will know about this HealthChecker.
// Panic if 'key' is already present.
func AddHealthChecker(key string, checker HealthChecker) {
	checkerLock.Lock()
	defer checkerLock.Unlock()
	if _, found := checkers[key]; found {
		glog.Fatalf("HealthChecker already defined for key %s.", key)
	}
	checkers[key] = checker
}

// NewHealthChecker creates a new HealthChecker wichi supports multiple types of liveness proves.
func NewHealthChecker() HealthChecker {
	checkerLock.Lock()
	defer checkerLock.Unlock()
	input := map[string]HealthChecker{}
	for key, value := range checkers {
		input[key] = value
	}
	return &muxHealthChecker{
		checkers: input,
	}
}

// muxHealthChecker bundles multiple implementations of HealthChecker of different types.
type muxHealthChecker struct {
	checkers map[string]HealthChecker
}

// HealthCheck delegates the health-checking of the container to one of the bundled implementations.
// It choose an implementation according to container.LivenessProbe.Type.
// If there is no matching health checker it returns Unknown, nil.
func (m *muxHealthChecker) HealthCheck(podFullName string, currentState api.PodState, container api.Container) (Status, error) {
	checker, ok := m.checkers[container.LivenessProbe.Type]
	if !ok || checker == nil {
		glog.Warningf("Failed to find health checker for %s %s", container.Name, container.LivenessProbe.Type)
		return Unknown, nil
	}
	return checker.HealthCheck(podFullName, currentState, container)
}

// findPortByName is a helper function to look up a port in container by name.
// Returns the HostPort if found, -1 if not found.
func findPortByName(container api.Container, portName string) int {
	for _, port := range container.Ports {
		if port.Name == portName {
			return port.HostPort
		}
	}
	return -1
}

func (s Status) String() string {
	if s == Healthy {
		return "healthy"
	} else if s == Unhealthy {
		return "unhealthy"
	}
	return "unknown"
}
