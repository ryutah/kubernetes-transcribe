package api

import (
	docker "github.com/fsouza/go-dockerclient"
	"github.com/ryutah/kubernetes-transcribe/pkg/util"
)

// Common string formats
// ---------------------
// Many fields in this API have formatting requirements.  The commonly used
// formats are defined here.
//
// C_IDENTIFIER:  This is a string that conforms the definition of an "identifier"
//     in the C language.  This is captured by the following regex:
//         [A-za-z_][A-za-z0-9_]*
//     This defines the format, but not the length restriction, which should be
//     specified at the definition of any field of this type.
//
// DNS_LABEL:  This is a string, no more than 63 caracters long, that conforms
//     to the definition of a "subdomain" in RFCs 1035 and 1123.  This is captured
//     by the following regex:
//         [a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*
//     or more simply:
//         DNS_LABEL(\.DNS_LABEL)*

// ContainerManifest corresponds to the Container Manifest format.
// Refs: https://developers.google.com/compute/docs/containers/container_vms#container_manifest
// This is used as the representation of Kubernetes workloads.
type ContainerManifest struct {
	// Required: This must be a supported version string, such as "v1beta1".
	Version string `json:"version" yaml:"version"`
	// Required: This must be a DNS_SUBDOMAIN.
	// TODO: ID on Manifest is deprecated and will be removed in the future.
	ID string `json:"id" yaml:"id"`
	// TODO: UUID on Manifest is deprecated in the future once we are done
	// with the API refactoring. It is required for now to determine the instance
	// of a Pod.
	UUID          string        `json:"uuid,omitempty" yaml:"uuid,omitempty"`
	Volumes       []Volume      `json:"volumes" yaml:"volumes"`
	Containers    []Container   `json:"containers" yaml:"containers"`
	RestartPolicy RestartPolicy `json:"restartPolicy,omitempty" yaml:"restartPolicy,omitempty"`
}

// ContainerManifestList is used to communicate container manifests to kubelet.
type ContainerManifestList struct {
	JSONBase `json:",inline" yaml:",inline"`
	Items    []ContainerManifest `json:"items,omitempty" yaml:"items,omitempty"`
}

func (*ContainerManifestList) IsAnAPIObject() {}

// Volume represents a named volume in a pod that may be accessed by any containers in the pod.
type Volume struct {
	// Required: This must be a DNS_LABEL. Each volume in a pod must have a unique name.
	Name string `json:"name" yaml:"name"`
	// Source represents the location and type of a volume to mount.
	// This is optional for now. If not specified, the Volume is implied to be an EmptyDirectory.
	// This implied behavior is deprecateed and will be removed in a future version.
	Source *VolumeSource `json:"source" yaml:"source"`
}

type VolumeSource struct {
	// Only one of the following sources may be specified
	// HostDirectory represents a pre-existing directory on the host machine that is directry
	// exposed to the container. This is generally used for system agents or other privileged
	// things that are allowd to see the host machine. Host containers will NOT need this.
	// TODO(jsonsdl) We need to restrict who can use host directory mounts and
	// who can/can not mount host directories as read/write.
	HostDirectory *HostDirectory `json:"hostDir" yaml:"hostDir"`
	// EmptyDirectory represents a temporary directory that share a pod's lifetime.
	EmptyDirectory *EmptyDirectory `json:"emptyDir" yaml:"emptyDir"`
}

// HostDirectory represents bare host directory volume.
type HostDirectory struct {
	Path string `json:"path" yaml:"path"`
}

type EmptyDirectory struct{}

// Port represents a network port in a single container.
type Port struct {
	// Optional: If specified, this must be a DNS_LABEL. Each named port in a pod must have a unique name.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Optional: If specified, this must be a valied port number, 0 < x < 65536.
	HostPort int `json:"hostPort,omitempty" yaml:"hostPort,omitempty"`
	// Required: This must be a valid port number, 0 < x < 65536.
	ContainerPort int `json:"containerPort" yaml:"containerPort"`
	// Optional: Supports "TCP" and "UDP". Defaults to "TCP"
	Protocol string `json:"protocol,omitempty" yaml:"protocol,omitempty"`
	// Optional: What host IP to bind the external port to.
	HostIP string `json:"hostIP,omitempty" yaml:"hostIP,omitempty"`
}

// VolumeMount describes a mounting of Volume within a container.
type VolumeMount struct {
	// Required: This must match the Name of a Volume [above].
	Name string `json:"name" yaml:"name"`
	// Optional: Defaults to false (read-write).
	ReadOnly bool `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	// Required.
	MountPath string `json:"mountPath,omitempty" yaml:"mountPath,omitempty"`
}

// EnvVar represents an environment variable present in a Container.
type EnvVar struct {
	// Required: This must be a C_IDENTIFIER.
	Name string `json:"name" yaml:"name"`
	// Optional: defaults to "".
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
}

// HTTPGetAction describes an action based on HTTP Get requests.
type HTTPGetAction struct {
	// Optional: Path to access on the HTTP server.
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
	// Required: Name or number of the port to access on the container.
	Port util.IntOrString `json:"port,omitempty" yaml:"port,omitempty"`
	// Optional: Host name to connect to, defaults to the pod IP.
	Host string `json:"host,omitempty" yaml:"host,omitempty"`
}

// TCPSocketAction describes an action based on opening a socket.
type TCPSocketAction struct {
	// Required: Port to connect to.
	Port util.IntOrString `json:"port,omitempty" yaml:"port,omitempty"`
}

// ExecAction describes a "run in container" action
type ExecAction struct {
	// Command is the command line to execute inside the container, the working directory for the
	// command is root ('/') in the container's filesystem.  The command is simply exec'd it is
	// not run inside a shell, so traditional shell instructions ('|', etc) won't work.
	// a shell, you need to explicitly call out to the shell.
	Command []string `json:"command,omitempty" yaml:"command,omitempty"`
}

// LivenessProbe describes a liveness probe to be examined to the container.
// TODO: pass structured data to the actions, and document that data here.
type LivenessProbe struct {
	Type                string           `json:"type,omitempty" yaml:"type,omitempty"`
	HTTPGet             *HTTPGetAction   `json:"httpGet,omitempty" yaml:"httpGet,omitempty"`
	TCPSocket           *TCPSocketAction `json:"tcpSocket,omitempty" yaml:"tcpSocket,omitempty"`
	Exec                *ExecAction      `json:"exec,omitempty" yaml:"exec,omitempty"`
	InitialDelaySeconds int64            `json:"initialDelaySeconds,omitempty" yaml:"initialDelaySeconds,omitempty"`
}

// Container represents a single container that is expected to be run on the host.
type Container struct {
	// Required: This must be a DNS_LABEL.  Each container in a pod must have a unique name.
	Name string
	// Required.
	Image string
	// Optional: Defaults to whatever is defined in the image.
	Command []string
	// Optional: Defaults to Docker's default.
	WorkingDir string
	Ports      []Port
	Env        []EnvVar
	// Optional: Defaults to unlimited.
	Memory int
	// Optional: Defaults to unlimited.
	CPU           int
	VolumeMounts  []VolumeMount
	LivenessProbe *LivenessProbe
	Lifecycle     *Lifecycle
	// Optional: Default to false.
	Privileged bool
}

// Handler defines a specific action that should be taken.
// TODO: pass structured data to these actions, and document that data here.
type Handler struct {
	// One add only one of the following should be specified.
	// Exec specifies the action to take.
	Exec *ExecAction `json:"exec,omitempty" yaml:"exec,omitempty"`
	// HTTPGet specifies the http request to perform.
	HTTPGet *HTTPGetAction `json:"httpGet,omitempty" yaml:"httpGet,omitempty"`
}

// Lifecycle describes actions that the management system should take in response to container lifecycle events.
// For the PostStart and PreStop lifecycle handlers, management of the container blocks unless the action is complete,
// unless the container process fails, in which case the handler is aborted.
type Lifecycle struct {
	// PostStart is called immediately after a container is created.
	// If the handler failds, the container is terminated and restarted.
	PostStart *Handler `json:"postStart,omitempty" yaml:"postStart,omitempty"`
	// PreStop is called immediately before a container is terminated.
	// The reason for termination is passed to the handle passed to the handler.
	// Regardless of the outcome of the handler, the container is eventually teminated.
	PreStop *Handler `json:"preStop,omitempty" yaml:"preStop,omitempty"`
}

// Event is the representation of an event logged to etcd backends.
type Event struct {
	Event     string             `json:"event,omitempty"`
	Manifest  *ContainerManifest `json:"manifest,omitempty"`
	Container *Container         `json:"container,omitempty"`
	Timestamp int64              `json:"timestamp"`
}

// The below types are used by kube_client and api_serer.

// JSONBase is shared by all objects sent to, or returnd from the client.
type JSONBase struct {
	Kind              string    `json:"kind,omitempty" yaml:"kind,omitempty"`
	ID                string    `json:"id,omitempty" yaml:"id,omitempty"`
	CreationTimestamp util.Time `json:"creationTimestamp,omitempty" yaml:"creationTimestamp,omitempty"`
	SelfLink          string    `json:"selfLink,omitempty" yaml:"selfLink,omitempty"`
	ResourceVersion   uint64    `json:"resourceVersion,omitempty" yaml:"resourceVersion,omitempty"`
	APIVersion        string    `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
}

// PodStatus represents a status of a pod.
type PodStatus string

// These are the valid statuses of pods.
const (
	PodWaiting    PodStatus = "Waiting"
	PodRunning    PodStatus = "Running"
	PodTerminated PodStatus = "Terminated"
)

type ContainerStateWaiting struct {
	Reason string `json:"reason,omitempty" yaml:"reason,omitempty"`
}

type ContainerStateRunning struct {
}

type ContainerStateTerminated struct {
	ExitCode int    `json:"exitCode,omitempty" yaml:"exitCode,omitempty"`
	Signal   int    `json:"signal,omitempty" yaml:"signal,omitempty"`
	Reason   string `json:"reason,omitempty" yaml:"reason,omitempty"`
}

type ContainerState struct {
	// Only one of the following ContainerState may be specified.
	// If none of them is specified, the default one is ContainerStateWaiting.
	Waiting     *ContainerStateWaiting    `json:"waiting,omitempty" yaml:"waiting,omitempty"`
	Running     *ContainerStateRunning    `json:"running,omitempty" yaml:"running,omitempty"`
	Termination *ContainerStateTerminated `json:"termination,omitempty" yaml:"termination,omitempty"`
}

type ContainerStatus struct {
	// TODO: Should we rename PodStatus to a more generic name or have a separate states.
	// defined for container?
	State        ContainerState `json:"state,omitempty" yaml:"state,omitempty"`
	RestartCount int            `json:"restartCount,omitempty" yaml:"restartCount,omitempty"`
	// the dependency on docker.
	DetailInfo docker.Container `json:"detailInfo,omitempty" yaml:"detailInfo,omitempty"`
}

type PodInfo map[string]docker.Container

type RestartPolicyAlways struct{}

// TODO: Define what kinds of failures should restart.
// TODO: Decide whether to support policy knobs, and, if so, which ones.
type RestartPolicyOnFailure struct{}

type RestartPolicyNever struct{}

type RestartPolicy struct {
	// Only one of the following restart policies may be specified.
	// If none of the following policies is specified, the default one is RestartPolicyAlways
	Always    *RestartPolicyAlways    `json:"always,omitempty" yaml:"always,omitempty"`
	OnFailure *RestartPolicyOnFailure `json:"onFailure,omitempty" yaml:"onFailure,omitempty"`
	Never     *RestartPolicyNever     `json:"never,omitempty" yaml:"never,omitempty"`
}

// PodState is the state of a pod, used as either input (desired stete) or output (current state).
type PodState struct {
	Manifest ContainerManifest `json:"manifest,omitempty" yaml:"manifest,omitempty"`
	Status   PodStatus         `json:"status,omitempty" yaml:"status,omitempty"`
	Host     string            `json:"host,omitempty" yaml:"host,omitempty"`
	HostIP   string            `json:"hostIP,omitempty" yaml:"hostIP,omitempty"`
	PodIP    string            `json:"podIP,omitempty" yaml:"podIP,omitempty"`

	// The key of this map is the *name* of the container within the manifest; it has one
	// entry per container in the manifest. The value of this map is currently the output
	// of `docker inspect`. This output format is *not* final and should not be relied upon.
	// TODO: Make real decisions about what out info should look like. Re-enable fuzz test
	// when we have done this.
	Info PodInfo `json:"info,omitempty" yaml:"info,omitempty"`
}

// PodList is a list of Pods.
type PodList struct {
	JSONBase `json:",inline" yaml:",inline"`
	Items    []Pod `json:"items" yaml:"items,omitempty"`
}

func (*PodList) IsAnAPIObject() {}

// Pod is a collection of containers, used as either input (create, update) or as ouput (list, get).
type Pod struct {
	JSONBase     `json:",inline" yaml:",inline"`
	Labels       map[string]string
	DesiredState PodState `json:"desiredState,omitempty" yaml:"desiredState,omitempty"`
	CurrentState PodState `json:"currentState,omitempty" yaml:"currentState,omitempty"`
}

func (*Pod) IsAnAPIObject() {}

// ReplicationControllerState is the state of a replication controller, eigher input (create, update) or as output (list, get).
type ReplicationControllerState struct {
	Replicas        int               `json:"replicas" yaml:"replicas"`
	ReplicaSelector map[string]string `json:"replicaSelector,omitempty" yaml:"replicaSelector,omitempty"`
	PodTemplate     PodTemplate       `json:"podTemplate,omitempty" yaml:"podTemplate,omitempty"`
}

// ReplicationControllerList is a collection of replication controllers.
type ReplicationControllerList struct {
	JSONBase `json:",inline" yaml:",inline"`
	Items    []ReplicationController `json:"items,omitempty" yaml:"items,omitempty"`
}

func (*ReplicationControllerList) IsAnAPIObject() {}

// ReplicationController represents the configuration of a replication controller.
type ReplicationController struct {
	JSONBase     `json:",inline" yaml:",inline"`
	DesiredState ReplicationControllerState `json:"desiredState,omitempty" yaml:"desiredState,omitempty"`
	CurrentState ReplicationControllerState `json:"currentState,omitempty" yaml:"currentState,omitempty"`
	Labels       map[string]string          `json:"labels,omitempty" yaml:"labels,omitempty"`
}

func (*ReplicationController) IsAnAPIObject() {}

// PodTemplate holds the information used for creating pods.
type PodTemplate struct {
	DesiredState PodState          `json:"desiredState,omitempty" yaml:"desiredState,omitempty"`
	Labels       map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
}

// ServiceList holds a list of services.
type ServiceList struct {
	JSONBase `json:",inline" yaml:",inline"`
	Items    []Service `json:"items" yaml:"items"`
}

func (*ServiceList) IsAnAPIObject() {}

// Service is a named abstraction of software service (for example, mysql) consisting of local port
// (for example 3306) that the proxy listens on, and the selector that determines which pors
// will answer requests sent through the proxy.
type Service struct {
	JSONBase `json:",inline" yaml:",inline"`

	// Required.
	Port int `json:"port" yaml:"port"`
	// Optional: Supports "TCP" and "UDP". Defaults to "TCP"
	Protocol string `json:"protocol" yaml:"protocol"`

	// This service's labels.
	Labels map[string]string `json:"labels" yaml:"labels"`

	// This service will route traffic to pods having labels matching this selector.
	Selector                   map[string]string `json:"selector" yaml:"selector"`
	CreateExternalLoadBalancer bool              `json:"create_external_load_balancer" yaml:"create_external_load_balancer"`

	// ContainerPort is the name of the port on the container to direct traffic to.
	// Optional, if unspecified use the first port on the container.
	ContanerPort util.IntOrString `json:"contaner_port" yaml:"contaner_port"`
}

func (*Service) IsAnAPIObject() {}

// Endpoints is a collection of a endpoints that implement the actual service, for example:
// Name: "mysql", Endpoints: ["10.10.1.1:1909", "10.10.2.2:8834"]
type Endpoints struct {
	JSONBase  `json:",inline" yaml:",inline"`
	Endpoints []string `json:"endpoints,omitempty" yaml:"endpoints,omitempty"`
}

func (*Endpoints) IsAnAPIObject() {}

// EndpointsList is a list of endpoints.
type EndpointsList struct {
	JSONBase `json:",inline" yaml:",inline"`
	Items    []Endpoints `json:"items,omitempty" yaml:"items,omitempty"`
}

func (*EndpointsList) IsAnAPIObject() {}

// Minion is a worker node in Kubernetes.
// The name of the minion according to etcd is in JSONBase.ID.
type Minion struct {
	JSONBase `json:",inline" yaml:",inline"`
	// Queried from cloud provider, if available.
	HostIP string `json:"hostIP,omitempty" yaml:"hostIP,omitempty"`
}

func (*Minion) IsAnAPIObject() {}

// MinionList is a list of minions.
type MinionList struct {
	JSONBase `json:",inline" yaml:",inline"`
	Items    []Minion `json:"items,omitempty" yaml:"items,omitempty"`
}

func (*MinionList) IsAnAPIObject() {}

// Binding is written by a scheduler to cause a pod to be bound to a host.
type Binding struct {
	JSONBase `json:",inline" yaml:",inline"`
	PodID    string `json:"podID" yaml:"podID"`
	Host     string `json:"host" yaml:"host"`
}

func (*Binding) IsAnAPIObject() {}

// Status is a return value for calls that don't return other objects.
// TODO: this could go in apiserver, but I'm including it here so clients needn't
// import both.
type Status struct {
	JSONBase `json:",inline" yaml:",inline"`
	// One of: "success", "failuer", "working" (for operations not yet completed)
	Status string `json:"status,omitempty" yaml:"status,omitempty"`
	// A human-readable description of the status of this operation.
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
	// A machine-readable description of why this operation is in the
	// "failure" or "working" status. If this value is empty there
	// is no information available. A Reason clarifies an HTTP status code
	// but does not override it.
	Reason StatusReason `json:"reason,omitempty" yaml:"reason,omitempty"`
	// Extended data associated with the reason.  Each reason may define its
	// own extended details. This field is optional and the data returned
	// is not guaranteed to conform to any schema except that defined by
	// the reason type.
	Details *StatusDetails `json:"details,omitempty" yaml:"details,omitempty"`
	// Suggested HTTP return code for this status, 0 if not set.
	Code int `json:"code,omitempty" yaml:"code,omitempty"`
}

func (*Status) IsAnAPIObject() {}

// StatusDetails is a set of additional properties that MAY be set by the
// server to provide additional information about a response. The Reason
// field of a Status object defines what attributes will be set. Clients
// must ignore fields that do not match the defined type of each attributes,
// defined.
type StatusDetails struct {
	// The ID attribute of the resource associated with the status StatusReason.
	// (when there is a single ID which can be described).
	ID string `json:"id,omitempty" yaml:"id,omitempty"`
	// The kind attribute of the resource associated with the status StatusReason.
	// On some operations may differ from the requested resources Kind.
	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`
	// The Cause array includes more details associated with the StatusReason
	// failure. Not all StatusReason may provide details causes.
	Causes []StatusCause `json:"causes,omitempty" yaml:"causes,omitempty"`
}

// Values of Status.Status
const (
	StatusSuccess = "success"
	StatusFailure = "failure"
	StatusWorking = "working"
)

// StatusReason is an enumeration of possible failure causes.  Each StatusReason
// must map to a single HTTP status code, but multiple reasons may map
// to the same HTTP status code.
// TODO: move to apiserver
type StatusReason string

const (
	// StatusReasonUnknown means the server has declined to indicate a specific reason.
	// The details field may contain other information about this error.
	// Status code 500.
	StatusReasonUnknown StatusReason = ""

	// StatusReasonWorking menas the server is processing this result and will complete
	// at a future time.
	// Details (optional):
	//   "kind" string - the name of the resource being referenced ("operation" today)
	//   "id"   string - the identifier of the Operation resource where updates
	//                   will be returned
	//
	// Headers (optional):
	//   "Location" - HTTP header populated with a URL that can retrieved the final
	//                status of this operation
	// Status code 202
	StatusReasonWorking StatusReason = "working"

	// StatusReasonNotFound means one or more resources required for this operation
	// could not be found.
	// Details (optional):
	//   "kind" string - the kind attribute of the missing resource
	//                   on some operations may differ from the requested
	//                   resource.
	//   "id"   string - the identifier of the missing resource
	// Status code 404
	StatusReasonNotFound StatusReason = "not_found"

	// StatusReasonAlreadyExists means the resource you are creating already exists.
	// Details (optional):
	//   "kind" string - the kind attribute of the conflicting resource
	//   "id"   string - the identifier of the conflicting resource
	// Status code 409
	StatusReasonAlreadyExists StatusReason = "already_exists"

	// StatusReasonConflict means the requested update operation cannot be completed
	// due to a conflict in the operation. The client may need to alter the request.
	// Each resource may define custom details that indicate the nature of the
	// conflict.
	// Status code 409
	StatusReasonConflict StatusReason = "conflict"

	// StatusReasonInvalid means the requested create or update operation cannot be
	// completed due to invalid data provided as part of the request. The client may
	// need to alter the request. When set, the client may use the StatusDetails
	// message field as a summary of the issues encountered.
	// Details (optional):
	//   "kind" string - the kind attribute of the invalid resource
	//   "id"   string - the identifier of the invalid resource
	//   "causes"      - one or more StatusCause entries indicating the data in the
	//                   provided resource that was invalid.  The code, message, and
	//                   field attributes will be set.
	// Status code 422
	StatusReasonInvalid StatusReason = "invalid"
)

// StatusCause provides more information about an api.Status failure, including
// cases when multiple errors are encounterd.
type StatusCause struct {
	// A machine-readable description of the cause of the error. If this value is
	// empty there is no information available.
	Type CauseType `json:"type,omitempty" yaml:"type,omitempty"`
	// A human-readable description of the cause of the error. This field may be
	// presented as-is to a reader.
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
	// The field of the resource that has caused this error, as named by its JSON
	// serialization. May include dot and postfix notation for nasted attributes.
	// Arrays are zero-indexed. Fields may appear more than once in an array of
	// causes due to fields having multiple errors.
	// Optional.
	//
	// Examples:
	//   "name": - the field "name" on the current resource
	//   "items[0].name" - the field "name" on the first array entry in "items"
	Field string `json:"field,omitempty" yaml:"field,omitempty"`
}

// CauseType is a machine readable value providing more detail about what
// occured in a status response. An operation may have multiple causes for a
// status (whether failure, success, or working)
type CauseType string

const (
	// CauseTypeFieldValueNotFound is used to report failure to find a requested value
	// (e.g. looking up an ID).
	CauseTypeFieldValueNotFound CauseType = "fieldValueNotFound"
	// CauseTypeFieldValueRequired is used to report required values that are not
	// provided (e.g. empty strings, null values, or empty arrays).
	CauseTypeFieldValueRequired CauseType = "fieldValueRequired"
	// CauseTypeFieldValueDuplicate is used to report collisions of values that must be
	// unique (e.g. unique IDs).
	CauseTypeFieldValueDuplicate CauseType = "fieldValueDuplicate"
	// CauseTypeFieldValueInvalid is used to report malformed values (e.g. failed regex
	// match).
	CauseTypeFieldValueInvalid CauseType = "fieldValueInvalid"
	// CauseTypeFieldValueNotSupported is used to report valid (as per formatting rules)
	// values that can not be handled (e.g. an enumerated string).
	CauseTypeFieldValueNotSupported CauseType = "fieldValueNotSupported"
)

// ServerOp is an operation delivered to API clients.
type ServerOp struct {
	JSONBase `json:",inline" yaml:",inline"`
}

func (*ServerOp) IsAnAPIObject() {}

// ServerOpList is a list of operations, as delivered to API clients.
type ServerOpList struct {
	JSONBase `json:",inline" yaml:",inline"`
	Items    []ServerOp `json:"items,omitempty" yaml:"items,omitempty"`
}

func (*ServerOpList) IsAnAPIObject() {}
