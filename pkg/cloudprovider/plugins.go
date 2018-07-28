package cloudprovider

import (
	"io"
	"sync"

	"github.com/golang/glog"
)

// Factory is a function that returns a cloudprovider.Interface.
// The config parameter provides an io.Reader handler to the factory in
// order to load specific configurations. If no configurations is provided
// the parameter is nil.
type Factory func(config io.Reader) (Interface, error)

// All registered cloud providers.
var providersMutex sync.Mutex
var providers = make(map[string]Factory)

// RegisterCloudProvider registers a cloudprovider.Factory by name.  This
// is expected to happend during app startup.
func RegisterCloudProvider(name string, cloud Factory) {
	providersMutex.Lock()
	defer providersMutex.Unlock()
	_, found := providers[name]
	if found {
		glog.Fatalf("Cloud provider %q was registered twice", name)
	}
	glog.Infof("Registered cloud provider %q", name)
	providers[name] = cloud
}

// GetCloudProvider creates an instance of the named cloud provider, or nil if
// the name is not known.  The error return is only used if the named provider
// was known but failed to initialize. The config parameter specifies the
// io.Reader handler of the configuration file for the cloud provider, or nil
// for no configuration.
func GetCloudProvider(name string, config io.Reader) (Interface, error) {
	providersMutex.Lock()
	defer providersMutex.Unlock()
	f, found := providers[name]
	if !found {
		return nil, nil
	}
	return f(config)
}
