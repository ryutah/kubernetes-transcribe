// Package verflag defines utility functions to handle command line flags
// related to version of Kubernetes.
package verflag

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/ryutah/kubernetes-transcribe/pkg/version"
)

type versionValue int

const (
	VersionFalse versionValue = 0
	VersionTrue  versionValue = 1
	VersionRaw   versionValue = 2
)

const strRawVersion string = "raw"

func (v *versionValue) IsBoolFlag() bool {
	return true
}

func (v *versionValue) Get() interface{} {
	return versionValue(*v)
}

func (v *versionValue) Set(s string) error {
	if s == strRawVersion {
		*v = VersionRaw
		return nil
	}
	boolVal, err := strconv.ParseBool(s)
	if boolVal {
		*v = VersionTrue
	} else {
		*v = VersionFalse
	}
	return err
}

func (v *versionValue) String() string {
	if *v == VersionRaw {
		return strRawVersion
	}
	return fmt.Sprintf("%v", bool(*v == VersionTrue))
}

func VersionVar(p *versionValue, name string, value versionValue, usage string) {
	*p = value
	flag.Var(p, name, usage)
}

func Version(name string, value versionValue, usage string) *versionValue {
	p := new(versionValue)
	VersionVar(p, name, value, usage)
	return p
}

var (
	versionFlag = Version("version", VersionFalse, "Print version information and quit")
)

// PrintAndExitIfRequested will check if the -version flag was passed
// and, if so, print the version and exit.
func PrintAndExitIfRequested() {
	if *versionFlag == VersionRaw {
		fmt.Printf("%#v\n", version.Get())
		os.Exit(0)
	} else if *versionFlag == VersionTrue {
		fmt.Printf("Kubernetes %s\n", version.Get())
		os.Exit(0)
	}
}
