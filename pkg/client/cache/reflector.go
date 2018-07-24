package cache

import (
	"fmt"
	"reflect"
	"time"

	"github.com/golang/glog"
	"github.com/ryutah/kubernetes-transcribe/pkg/runtime"
	"github.com/ryutah/kubernetes-transcribe/pkg/util"
	"github.com/ryutah/kubernetes-transcribe/pkg/watch"
)

// ListerWatcher is any object that knows how to perform an initial list and start watch on a resource.
type ListerWatcher interface {
	// List should return a list type object; the Items field will be extracted, and the
	// ResourceVersion field will be used to start the watch in the right place.
	List() (runtime.Object, error)
	// Watch should begin a watch at the specified version.
	Watch(resourceVersion uint64) (watch.Interface, error)
}

// Reflector watches a specified resource and causes all changes to be reflected in the given store.
type Reflector struct {
	// The type of object we expected to place in the store.
	expectedType reflect.Type
	// The destination to sync up with the watch source.
	store Store
	// listerWatcher is used to perform lists and watches.
	listerWatcher ListerWatcher
	// period controls timing between one watch ending and
	// the beginning of the next one.
	period time.Duration
}

// Run starts a watch and handles watch events. Will restart the watch if it is closed.
// Run starts a goroutine and returns immediately.
func (r *Reflector) Run() {
	go util.Forever(func() {
		r.listAndWatch()
	}, r.period)
}

func (r *Reflector) listAndWatch() {
	var resourceVersion uint64

	list, err := r.listerWatcher.List()
	if err != nil {
		glog.Errorf("Failed to list %v: %v", r.expectedType, err)
		return
	}
	jsonBase, err := runtime.FindJSONBase(list)
	if err != nil {
		glog.Errorf("Unable to understand list result %#v", list)
		return
	}
	resourceVersion = jsonBase.ResourceVersion()
	items, err := runtime.ExtractList(list)
	if err != nil {
		glog.Errorf("Unable to understand list result %#v (%v)", list, err)
		return
	}
	err = r.syncWith(items)
	if err != nil {
		glog.Errorf("Unable to sync list result: %v", err)
		return
	}

	for {
		w, err := r.listerWatcher.Watch(resourceVersion)
		if err != nil {
			glog.Errorf("failed to watch %v: %v", r.expectedType, err)
			return
		}
		r.watchHandler(w, &resourceVersion)
	}
}

// syncWith replace the store's items with the given list.
func (r *Reflector) syncWith(items []runtime.Object) error {
	found := map[string]interface{}{}
	for _, item := range items {
		jsonBase, err := runtime.FindJSONBase(item)
		if err != nil {
			return fmt.Errorf("unexpected item in list: %v", err)
		}
		found[jsonBase.ID()] = item
	}

	r.store.Replace(found)
	return nil
}

// watchHandler watches w and keep *resourceVersion up to date.
func (r *Reflector) watchHandler(w watch.Interface, resourceVersion *uint64) {
	for {
		event, ok := <-w.ResultChan()
		if !ok {
			glog.Errorf("unexpected watch close")
			return
		}
		if e, a := r.expectedType, reflect.TypeOf(event.Object); e != a {
			glog.Errorf("expected type %v, but watch event object had type %v", e, a)
			continue
		}
		jsonBase, err := runtime.FindJSONBase(event.Object)
		if err != nil {
			glog.Errorf("unable to understand watch event %#v", event)
			continue
		}
		switch event.Type {
		case watch.Added:
			r.store.Add(jsonBase.ID(), event.Object)
		case watch.Modified:
			r.store.Update(jsonBase.ID(), event.Object)
		case watch.Deleted:
			// TODO: Will any consumers need access to the "last known
			// state", which is passed in event.Object? If so, may need
			// to change this.
			r.store.Delete(jsonBase.ID())
		default:
			glog.Errorf("unable to understand watch event %#v", event)
		}
		*resourceVersion = jsonBase.ResourceVersion() + 1
	}
}
