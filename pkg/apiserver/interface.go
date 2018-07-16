package apiserver

import (
	"github.com/ryutah/kubernetes-transcribe/pkg/labels"
	"github.com/ryutah/kubernetes-transcribe/pkg/runtime"
	"github.com/ryutah/kubernetes-transcribe/pkg/watch"
)

// RESTStorage is a generig interface for RESTful storage services.
// Resources which are exported to the RESTful API or apiserver need to implement this interface.
type RESTStorage interface {
	// New returns an empty object that can be used with Create and Update after request data has been put into it.
	// This object must be a pointer type for use with Codec.DecodeInto([]byte, runtime.Object)
	New() runtime.Object

	// List selects resources in the storage which match to the selector.
	List(label, field labels.Selector) (runtime.Object, error)

	// Get finds a resource in the storage by id and returns it.
	// Although it can return an arbitrary erorr value, IsNotFound(err) is true for the
	// returned error value err when the specified resouce is not found.
	Get(id string) (runtime.Object, error)

	// Delete finds a resource in the storage and deletes it.
	// Although it can return an arbitrary erorr value, IsNotFound(err) is true for the
	// returned error value err when the specified resouce is not found.
	Delete(id string) (<-chan runtime.Object, error)

	Create(runtime.Object) (<-chan runtime.Object, error)
	Update(runtime.Object) (<-chan runtime.Object, error)
}

// ResourceWatcher should be implemented by all RESTStorage object that
// want to offer the abillity to watch for changes through watch api.
type ResourceWatcher interface {
	// 'label' selects on labels; 'field' selects on the object's fields. Not all fields are supported;
	// an error should be returned if 'field' tries to select on a field that
	// is't supported 'resourceVersion' allows for continuing/starting a watch at a particular version.
	Watch(label, field labels.Selector, resourceVersion uint64) (watch.Interface, error)
}

// Redirector know how to return a remote resource's location.
type Redirector interface {
	// ResourceLocation should return the remote location of the given resource, or an error.
	ResourceLocation(id string) (remoteLocation string, err error)
}
