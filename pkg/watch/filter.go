package watch

// FilterFunc should take an event, possibly modify it in some way, and return
// the modified event. If the event should be ignored, then result keep=false.
type FilterFunc func(in Event) (out Event, keep bool)

// Filter passes all events through f before allowing them to pass on.
// Putting a filter on a watch, as an unavoidable side-efffect due to the way
// go channels work, effectively causes the watch's event channel to have its
// queue length increased by one.
//
// WARNING: filter has a fatal flaw, in that it can't properly update the
// type field (Add/Modified/Deleted) to reflect items beginning to pass the
// filter when they previously didn't
func Filter(w Interface, f FilterFunc) Interface {
	fw := &filterWatch{
		incoming: w,
		result:   make(chan Event),
		f:        f,
	}
	go fw.loop()
	return fw
}

type filterWatch struct {
	incoming Interface
	result   chan Event
	f        FilterFunc
}

// ResultChan returns a channel which will receive filtered events.
func (fw *filterWatch) ResultChan() <-chan Event {
	return fw.result
}

// Stop stops the upstream watch, which will eventually stop this watch.
func (fw *filterWatch) Stop() {
	fw.incoming.Stop()
}

// loop waits for new values, filters them, and resends them.
func (fw *filterWatch) loop() {
	defer close(fw.result)
	for {
		event, ok := <-fw.incoming.ResultChan()
		if !ok {
			break
		}
		filtered, keep := fw.f(event)
		if keep {
			fw.result <- filtered
		}
	}
}
