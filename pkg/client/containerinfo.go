package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"

	"github.com/google/cadvisor/info"
)

type ContainerInfoGetter interface {
	// GetContainerInfo returns information about a container.
	GetContainerInfo(host, podID, containerID string, req *info.ContainerInfoRequest) (*info.ContainerInfo, error)
	// GetRootInfo returns information about the root container on a machine.
	GetRootInfo(host string, req *info.ContainerInfoRequest) (*info.ContainerInfo, error)
	// GetMachineInfo returns the machine's information like number of cores, memory capacity.
	GetMachineInfo(host string) (*info.MachineInfo, error)
}

type HTTPContainerInfoGetter struct {
	Client *http.Client
	Port   int
}

func (h *HTTPContainerInfoGetter) GetMachineInfo(host string) (*info.MachineInfo, error) {
	request, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"http://%v/spec",
			net.JoinHostPort(host, strconv.Itoa(h.Port)),
		),
		nil,
	)
	if err != nil {
		return nil, err
	}

	response, err := h.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"trying to get machine spec from %v; received status %v",
			host, response.Status,
		)
	}
	var minfo info.MachineInfo
	err = json.NewDecoder(response.Body).Decode(&minfo)
	if err != nil {
		return nil, err
	}
	return &minfo, nil
}

func (h *HTTPContainerInfoGetter) getContainerInfo(host, path string, req *info.ContainerInfoRequest) (*info.ContainerInfo, error) {
	var body io.Reader
	if req != nil {
		content, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(content)
	}

	request, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"http://%v/stats/%v",
			net.JoinHostPort(host, strconv.Itoa(h.Port)),
			path,
		),
		body,
	)
	if err != nil {
		return nil, err
	}

	response, err := h.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"trying to get info for %v from %v; received status %v",
			path, host, response.Status,
		)
	}
	var cinfo info.ContainerInfo
	err = json.NewDecoder(response.Body).Decode(&cinfo)
	if err != nil {
		return nil, err
	}
	return &cinfo, nil
}

func (h *HTTPContainerInfoGetter) GetContainerInfo(host string, podID string, containerID string, req *info.ContainerInfoRequest) (*info.ContainerInfo, error) {
	return h.getContainerInfo(
		host,
		fmt.Sprintf("%v/%v", podID, containerID),
		req,
	)
}

func (h *HTTPContainerInfoGetter) GetRootInfo(host string, req *info.ContainerInfoRequest) (*info.ContainerInfo, error) {
	return h.getContainerInfo(host, "", req)
}
