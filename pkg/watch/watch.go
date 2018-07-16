package watch

import (
	"sync"

	"github.com/ryutah/kubernetes-transcribe/pkg/runtime"
)

// Interface can be implementes by anything that knows how to watch and report changes.
type Interface interface {
	// Stops watching, Will close the channel returned by ResultChan(). Releases
	// any resources used by the watch.
	Stop()

	// Returns a chan which will receive all the events. If an error occures
	// or Stop() is called, this channel will be closed, in which case the
	// watch should be completely cleaned up.
	ResultChan() <-chan Event
}

// EventType defines the possible types of events
type EventType string

const (
	Added    EventType = "ADDED"
	Modified EventType = "MODIFIED"
	Deleted  EventType = "DELETED"
)

// Event represents a single event to a watched resource.
type Event struct {
	Type EventType

	// If Type == Deleted, then this is the state of the object
	// immediately before deletion.
	Object runtime.Object
}

// FakeWatcher lets you test anything that consumes a watch.Interface; threadsafe.
type FakeWatcher struct {
	result  chan Event
	Stopped bool
	sync.Mutex
}

func (f *FakeWatcher) Stop() {
	f.Lock()
	defer f.Unlock()
	if !f.Stopped {
		close(f.result)
		f.Stopped = true
	}
}

func (f *FakeWatcher) ResultChan() <-chan Event {
	return f.result
}

// Add sends an add event.
func (f *FakeWatcher) Add(obj runtime.Object) {
	f.result <- Event{
		Type:   Added,
		Object: obj,
	}
}

// Modify sends a modify event.
func (f *FakeWatcher) Modify(obj runtime.Object) {
	f.result <- Event{
		Type:   Modified,
		Object: obj,
	}
}

// Delete sends a delete events.
func (f *FakeWatcher) Delete(lastValue runtime.Object) {
	f.result <- Event{
		Type:   Deleted,
		Object: lastValue,
	}
}

// Action sends an event of the requested type, for table-based testing.
func (f *FakeWatcher) Action(action EventType, obj runtime.Object) {
	f.result <- Event{
		Type:   action,
		Object: obj,
	}
}
