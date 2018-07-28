package health

import (
	"fmt"
	"net"
	"strconv"

	"github.com/golang/glog"
	"github.com/ryutah/kubernetes-transcribe/pkg/api"
	"github.com/ryutah/kubernetes-transcribe/pkg/util"
)

type TCPHealthChecker struct{}

// getTCPAddrParts parses the components of a TCP connection address.  For testability.
func getTCPAddrParts(currentState api.PodState, container api.Container) (string, int, error) {
	params := container.LivenessProbe.TCPSocket
	if params == nil {
		return "", -1, fmt.Errorf("error, no TCP parameters specified: %v", container)
	}
	port := -1
	switch params.Port.Kind {
	case util.IntstrInt:
		port = params.Port.IntVal
	case util.IntstrString:
		port = findPortByName(container, params.Port.StrVal)
		if port == -1 {
			// Last ditch effort - maybe it was an int stored as string?
			var err error
			if port, err = strconv.Atoi(params.Port.StrVal); err != nil {
				return "", -1, err
			}
		}
	}
	if port == -1 {
		return "", -1, fmt.Errorf("unknown port: %v", params.Port)
	}
	if len(currentState.PodIP) == 0 {
		return "", -1, fmt.Errorf("no host specified.")
	}

	return currentState.PodIP, port, nil
}

// DoTCPCheck checks that a TCP socket to the address can be opend.
// If the socket can be opened, it returns Healthy.
// If the socket failes to open, it returns Unhealthy.
// This is exported because some other packages want to do direct TCP checks.
func DoTCPCheck(addr string) (Status, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return Unhealthy, nil
	}
	err = conn.Close()
	if err != nil {
		glog.Errorf("unexpected error closing health check socket: %v (%#v)", err, err)
	}
	return Healthy, nil
}

func (t *TCPHealthChecker) HealthCheck(podFullName string, currentState api.PodState, container api.Container) (Status, error) {
	host, port, err := getTCPAddrParts(currentState, container)
	if err != nil {
		return Unknown, err
	}
	return DoTCPCheck(net.JoinHostPort(host, strconv.Itoa(port)))
}
