package api

import (
	"github.com/ryutah/kubernetes-transcribe/pkg/runtime"
)

var Scheme = runtime.NewScheme()

func init() {
	Scheme.AddKnownTypes("",
		&PodList{},
		&Pod{},
		&ReplicationControllerList{},
		&ReplicationController{},
		&ServiceList{},
		&Service{},
		&MinionList{},
		&Minion{},
		&Status{},
		&ServerOpList{},
		&ServerOp{},
		&ContainerManifestList{},
		&Endpoints{},
		&EndpointsList{},
		&Binding{},
	)
}
