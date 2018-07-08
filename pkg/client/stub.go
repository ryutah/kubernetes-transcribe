package client

import "net/http"

type (
	HTTPPodInfoGetter struct {
		Client *http.Client
		Port   uint
	}
	AuthInfo struct{}

	Client struct{}
)

type PodInfoGetter interface{}

func New(host, version string, auth *AuthInfo) (*Client, error) {
	panic("Not implement yet")
}
