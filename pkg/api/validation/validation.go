package validation

import (
	"strings"

	"github.com/ryutah/kubernetes-transcribe/pkg/api"
	errs "github.com/ryutah/kubernetes-transcribe/pkg/api/errors"
	"github.com/ryutah/kubernetes-transcribe/pkg/capabilities"
	"github.com/ryutah/kubernetes-transcribe/pkg/labels"
	"github.com/ryutah/kubernetes-transcribe/pkg/util"
)

func validateVolumes(volumes []api.Volume) (util.StringSet, errs.ErrorList) {
	allErrs := errs.ErrorList{}

	allNames := util.StringSet{}
	for i := range volumes {
		vol := &volumes[i] // so we can set default values
		el := errs.ErrorList{}
		// TODO: enforce that a source is set once we deprecate the impiled form.
		if vol.Source != nil {
			el = validateSource(vol.Source).Prefix("source")
		}
		if len(vol.Name) == 0 {
			el = append(el, errs.NewFieldRequired("name", vol.Name))
		} else if !util.IsDNSLabel(vol.Name) {
			el = append(el, errs.NewFieldInvalid("name", vol.Name))
		} else if allNames.Has(vol.Name) {
			el = append(el, errs.NewFieldDuplicate("name", vol.Name))
		}
		if len(el) == 0 {
			allNames.Insert(vol.Name)
		} else {
			allErrs = append(allErrs, el.PrefixIndex(i)...)
		}
	}
	return allNames, allErrs
}

var supportedPortProtocols = util.NewStringSet("TCP", "UDP")

func validatePorts(ports []api.Port) errs.ErrorList {
	allErrs := errs.ErrorList{}

	allNames := util.StringSet{}
	for i := range ports {
		pErrs := errs.ErrorList{}
		port := &ports[i] // so we can set default values
		if len(port.Name) > 0 {
			if len(port.Name) > 63 || !util.IsDNSLabel(port.Name) {
				pErrs = append(pErrs, errs.NewFieldInvalid("name", port.Name))
			} else if allNames.Has(port.Name) {
				pErrs = append(pErrs, errs.NewFieldDuplicate("name", port.Name))
			} else {
				allNames.Insert(port.Name)
			}
		}
		if port.ContainerPort == 0 {
			pErrs = append(pErrs, errs.NewFieldRequired("containerPort", port.ContainerPort))
		} else if !util.IsValidPortNum(port.ContainerPort) {
			pErrs = append(pErrs, errs.NewFieldInvalid("containerPort", port.ContainerPort))
		}
		if port.HostPort != 0 && !util.IsValidPortNum(port.HostPort) {
			pErrs = append(pErrs, errs.NewFieldInvalid("hostPort", port.HostPort))
		}
		if len(port.Protocol) == 0 {
			port.Protocol = "TCP"
		} else if !supportedPortProtocols.Has(strings.ToUpper(port.Protocol)) {
			pErrs = append(pErrs, errs.NewFieldNotSupported("protocol", port.Protocol))
		}
		allErrs = append(allErrs, pErrs.PrefixIndex(i)...)
	}
	return allErrs
}

func validateEnv(vars []api.EnvVar) errs.ErrorList {
	allErrs := errs.ErrorList{}

	for i := range vars {
		vErrs := errs.ErrorList{}
		ev := &vars[i] // so we can set default values
		if len(ev.Name) == 0 {
			vErrs = append(vErrs, errs.NewFieldRequired("name", ev.Name))
		}
		if !util.IsCIdentfier(ev.Name) {
			vErrs = append(vErrs, errs.NewFieldInvalid("name", ev.Name))
		}
		allErrs = append(allErrs, vErrs.PrefixIndex(i)...)
	}
	return allErrs
}

func validateVolumeMounts(mounts []api.VolumeMount, volumes util.StringSet) errs.ErrorList {
	allErrs := errs.ErrorList{}

	for i := range mounts {
		mErrs := errs.ErrorList{}
		mnt := &mounts[i] // so we can set default values
		if len(mnt.Name) == 0 {
			mErrs = append(mErrs, errs.NewFieldRequired("name", mnt.Name))
		} else if !volumes.Has(mnt.Name) {
			mErrs = append(mErrs, errs.NewNotFound("name", mnt.Name))
		}
		if len(mnt.MountPath) == 0 {
			mErrs = append(mErrs, errs.NewFieldRequired("mountPath", mnt.MountPath))
		}
		allErrs = append(allErrs, mErrs.PrefixIndex(i)...)
	}
	return allErrs
}

// AccumulateUniquePorts runs an extraction function on each Port of each Container,
// accumulating the results and returning an error if any ports conflict.
func AccumulateUniquePorts(containers []api.Container, accumulator map[int]bool, extract func(*api.Port) int) errs.ErrorList {
	allErrs := errs.ErrorList{}

	for ci := range containers {
		cErrs := errs.ErrorList{}
		ctr := &containers[ci]
		for pi := range ctr.Ports {
			port := extract(&ctr.Ports[pi])
			if port == 0 {
				continue
			}
			if accumulator[port] {
				cErrs = append(cErrs, errs.NewFieldDuplicate("Port", port))
			} else {
				accumulator[port] = true
			}
		}
		allErrs = append(allErrs, cErrs.PrefixIndex(ci)...)
	}
	return allErrs
}

// checkHostPortConflicts check for collidng Port.HostPort values across
// a slice of containers.
func checkHostPortConflicts(containers []api.Container) errs.ErrorList {
	allPorts := map[int]bool{}
	return AccumulateUniquePorts(containers, allPorts, func(p *api.Port) int { return p.HostPort })
}

func validateSource(source *api.VolumeSource) errs.ErrorList {
	numVolumes := 0
	allErrs := errs.ErrorList{}
	if source.HostDirectory != nil {
		numVolumes++
		allErrs = append(allErrs, validateHostDir(source.HostDirectory).Prefix("hostDirectory")...)
	}
	if source.EmptyDirectory != nil {
		numVolumes++
		// EmptyDirectory have nothing to validate
	}
	if numVolumes != 1 {
		allErrs = append(allErrs, errs.NewFieldInvalid("", source))
	}
	return allErrs
}

func validateHostDir(hostDir *api.HostDirectory) errs.ErrorList {
	allErrs := errs.ErrorList{}
	if hostDir.Path == "" {
		allErrs = append(allErrs, errs.NewNotFound("path", hostDir.Path))
	}
	return allErrs
}

func validateExecAction(exec *api.ExecAction) errs.ErrorList {
	allErrors := errs.ErrorList{}
	if len(exec.Command) == 0 {
		allErrors = append(allErrors, errs.NewFieldRequired("command", exec.Command))
	}
	return allErrors
}

func validateHTTPGetAction(http *api.HTTPGetAction) errs.ErrorList {
	allErrors := errs.ErrorList{}
	if len(http.Path) == 0 {
		allErrors = append(allErrors, errs.NewFieldRequired("path", http.Path))
	}
	return allErrors
}

func validateHandler(handler *api.Handler) errs.ErrorList {
	allErrs := errs.ErrorList{}
	if handler.Exec != nil {
		allErrs = append(allErrs, validateExecAction(handler.Exec).Prefix("exec")...)
	} else if handler.HTTPGet != nil {
		allErrs = append(allErrs, validateHTTPGetAction(handler.HTTPGet).Prefix("httpGet")...)
	} else {
		allErrs = append(allErrs, errs.NewFieldInvalid("", handler))
	}
	return allErrs
}

func validateLifecycle(lifecycle *api.Lifecycle) errs.ErrorList {
	allErrs := errs.ErrorList{}
	if lifecycle.PostStart != nil {
		allErrs = append(allErrs, validateHandler(lifecycle.PostStart).Prefix("postStart")...)
	}
	if lifecycle.PreStop != nil {
		allErrs = append(allErrs, validateHandler(lifecycle.PreStop).Prefix("preStop")...)
	}
	return allErrs
}

func validateContainers(containers []api.Container, volumes util.StringSet) errs.ErrorList {
	allErrs := errs.ErrorList{}

	allNames := util.StringSet{}
	for i := range containers {
		cErrs := errs.ErrorList{}
		ctr := &containers[i] // so we can set default values
		capabilities := capabilities.Get()
		if len(ctr.Name) == 0 {
			cErrs = append(cErrs, errs.NewFieldRequired("name", ctr.Name))
		} else if !util.IsDNSLabel(ctr.Name) {
			cErrs = append(cErrs, errs.NewFieldInvalid("name", ctr.Name))
		} else if allNames.Has(ctr.Name) {
			cErrs = append(cErrs, errs.NewFieldDuplicate("name", ctr.Name))
		} else if ctr.Privileged && !capabilities.AllowPrivileged {
			cErrs = append(cErrs, errs.NewFieldInvalid("privileged", ctr.Privileged))
		} else {
			allNames.Insert(ctr.Name)
		}
		if len(ctr.Image) == 0 {
			cErrs = append(cErrs, errs.NewFieldRequired("image", ctr.Image))
		}
		if ctr.Lifecycle != nil {
			cErrs = append(cErrs, validateLifecycle(ctr.Lifecycle).Prefix("lifecycle")...)
		}
		cErrs = append(cErrs, validatePorts(ctr.Ports).Prefix("ports")...)
		cErrs = append(cErrs, validateEnv(ctr.Env).Prefix("env")...)
		cErrs = append(cErrs, validateVolumeMounts(ctr.VolumeMounts, volumes).Prefix("volumeMounts")...)
		allErrs = append(allErrs, cErrs.PrefixIndex(i)...)
	}
	// Check for collidng ports across all containers.
	// TODO: This really is dependent on the network config of the host (IP per pod?)
	// and the config of the new manifest.  But we have not specced that out yet, so we'll just
	// make some assumptions for now.  As of now, pods share a network namespace, which means that
	// every Port.HostPort across the whole pod must be unique.
	allErrs = append(allErrs, checkHostPortConflicts(containers)...)

	return allErrs
}

var supportedManifestVersions = util.NewStringSet("v1beta1", "v1beta2")

// ValidateManifest tests that the specified ContainerManifest has valid data.
// This includes formatting and uniquness.  It also canonicalizes the
// structure by setting default values and implementing any backwords-compatibility
// tricks.
func ValidateManifest(manifest *api.ContainerManifest) errs.ErrorList {
	allErrs := errs.ErrorList{}

	if len(manifest.Version) == 0 {
		allErrs = append(allErrs, errs.NewFieldRequired("version", manifest.Version))
	} else if !supportedManifestVersions.Has(strings.ToLower(manifest.Version)) {
		allErrs = append(allErrs, errs.NewFieldNotSupported("version", manifest.Version))
	}
	allVolumes, vErrs := validateVolumes(manifest.Volumes)
	allErrs = append(allErrs, vErrs.Prefix("volumes")...)
	allErrs = append(allErrs, validateContainers(manifest.Containers, allVolumes).Prefix("container")...)
	allErrs = append(allErrs, validateRestartPolicy(&manifest.RestartPolicy).Prefix("restartPolicy")...)
	return allErrs
}

func validateRestartPolicy(restartPolicy *api.RestartPolicy) errs.ErrorList {
	numPolicies := 0
	allErrors := errs.ErrorList{}
	if restartPolicy.Always != nil {
		numPolicies++
	}
	if restartPolicy.OnFailure != nil {
		numPolicies++
	}
	if restartPolicy.Never != nil {
		numPolicies++
	}
	if numPolicies == 0 {
		restartPolicy.Always = &api.RestartPolicyAlways{}
	}
	if numPolicies > 1 {
		allErrors = append(allErrors, errs.NewFieldInvalid("", restartPolicy))
	}
	return allErrors
}

func ValidatePodState(podState *api.PodState) errs.ErrorList {
	allErrs := errs.ErrorList(ValidateManifest(&podState.Manifest)).Prefix("manifest")
	return allErrs
}

// ValidatePod tests if required fields in the pod are set.
func ValidatePod(pod *api.Pod) errs.ErrorList {
	allErrs := errs.ErrorList{}
	if len(pod.ID) == 0 {
		allErrs = append(allErrs, errs.NewFieldRequired("id", pod.ID))
	}
	allErrs = append(allErrs, ValidatePodState(&pod.DesiredState).Prefix("desiredState")...)
	return allErrs
}

// ValidateService tests if required fields in the service are set.
func ValidateService(service *api.Service) errs.ErrorList {
	allErrs := errs.ErrorList{}
	if len(service.ID) == 0 {
		allErrs = append(allErrs, errs.NewFieldRequired("id", service.ID))
	} else if !util.IsDNS952Label(service.ID) {
		allErrs = append(allErrs, errs.NewFieldInvalid("id", service.ID))
	}
	if !util.IsValidPortNum(service.Port) {
		allErrs = append(allErrs, errs.NewFieldInvalid("Service.Port", service.Port))
	}
	if len(service.Protocol) == 0 {
		service.Protocol = "TCP"
	} else if !supportedPortProtocols.Has(strings.ToUpper(service.Protocol)) {
		allErrs = append(allErrs, errs.NewFieldNotSupported("protocol", service.Protocol))
	}
	if labels.Set(service.Selector).AsSelector().Empty() {
		allErrs = append(allErrs, errs.NewFieldRequired("selector", service.Selector))
	}
	return allErrs
}

// ValidateReplicationController tests if required fields in the replication controller are set.
func ValidateReplicationController(controller *api.ReplicationController) errs.ErrorList {
	allErrs := errs.ErrorList{}
	if len(controller.ID) == 0 {
		allErrs = append(allErrs, errs.NewFieldRequired("id", controller.ID))
	}
	allErrs = append(allErrs, ValidateReplicationControllerState(&controller.DesiredState).Prefix("desiredState")...)
	return allErrs
}

// ValidateReplicationControllerState tests if required fields in the replication controller state are set.
func ValidateReplicationControllerState(state *api.ReplicationControllerState) errs.ErrorList {
	allErrs := errs.ErrorList{}
	if labels.Set(state.ReplicaSelector).AsSelector().Empty() {
		allErrs = append(allErrs, errs.NewFieldRequired("replicaSelector", state.ReplicaSelector))
	}
	selector := labels.Set(state.ReplicaSelector).AsSelector()
	labels := labels.Set(state.PodTemplate.Labels)
	if !selector.Matches(labels) {
		allErrs = append(allErrs, errs.NewFieldInvalid("podTemplate.labels", state.PodTemplate))
	}
	if state.Replicas < 0 {
		allErrs = append(allErrs, errs.NewFieldInvalid("replicas", state.Replicas))
	}
	allErrs = append(allErrs, ValidateManifest(&state.PodTemplate.DesiredState.Manifest).Prefix("podTemplate.desiredState.manifest")...)
	return allErrs
}
