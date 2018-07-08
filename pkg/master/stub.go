package master

import (
	"time"

	"github.com/ryutah/kubernetes-transcribe/pkg/apiserver"
	"github.com/ryutah/kubernetes-transcribe/pkg/client"
	"github.com/ryutah/kubernetes-transcribe/pkg/cloudprovider"
	"github.com/ryutah/kubernetes-transcribe/pkg/runtime"
	"github.com/ryutah/kubernetes-transcribe/pkg/tools"
)

type Master struct{}

func (m *Master) API_v1beta1() (map[string]apiserver.RESTStorage, runtime.Codec) {
	panic("Not implement yet")
}

func (m *Master) API_v1beta2() (map[string]apiserver.RESTStorage, runtime.Codec) {
	panic("Not implement yet")
}

type Config struct {
	Client             *client.Client
	Cloud              cloudprovider.Interface
	EtcdHelper         tools.EtcdHelper
	HealthCheckMinions bool
	Minions            []string
	MinionCacheTTL     time.Duration
	MinionRegexp       string
	PodInfoGetter      client.PodInfoGetter
}

func New(c *Config) *Master {
	panic("Not implement yet")
}

func NewEtcdHelper(etcdServers []string, version string) (helper tools.EtcdHelper, err error) {
	panic("Not implement yet")
}
