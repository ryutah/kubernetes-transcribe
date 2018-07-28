package apiserver

import (
	"github.com/ryutah/kubernetes-transcribe/pkg/runtime"
	"github.com/ryutah/kubernetes-transcribe/pkg/util"
)

// WorkFunc is used to perform any time consuming work for an api call, after
// the input has been validated. Pass one of these to MakeAsync to create an
// appropriate return value for the Update, Delete, and Create methods.
type WorkFunc func() (result runtime.Object, err error)

// MakeAsync takes a function and execute it, delivering the result in the way required
// by RESTStorage's Update, Delete and Create methods.
func MakeAsync(fn WorkFunc) <-chan runtime.Object {
	channel := make(chan runtime.Object)
	go func() {
		defer util.HandleCrash()
		obj, err := fn()
		if err != nil {
			channel <- errToAPIStatus(err)
		} else {
			channel <- obj
		}
		// 'close' is used to signal that no further values will
		// be written to the channel. Not strictly necessary, but
		// also won't hurt.
		close(channel)
	}()
	return channel
}
