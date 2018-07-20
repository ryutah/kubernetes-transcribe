package watch

import (
	"io"
	"reflect"
	"testing"

	"github.com/ryutah/kubernetes-transcribe/pkg/runtime"
)

type fakeDecoder struct {
	items chan Event
}

func (f fakeDecoder) Decode() (action EventType, object runtime.Object, err error) {
	item, open := <-f.items
	if !open {
		return action, nil, io.EOF
	}
	return item.Type, item.Object, nil
}

func (f fakeDecoder) Close() {
	close(f.items)
}

func TestStreamWatcher(t *testing.T) {
	table := []Event{
		{Added, testType("foo")},
	}

	fd := fakeDecoder{make(chan Event, 5)}
	sw := NewStreamWatcher(fd)

	for _, item := range table {
		fd.items <- item
		got, open := <-sw.ResultChan()
		if !open {
			t.Errorf("unexpected early close")
		}
		if e, a := item, got; !reflect.DeepEqual(e, a) {
			t.Errorf("expected %v, got %v", e, a)
		}
	}

	sw.Stop()
	_, open := <-sw.ResultChan()
	if open {
		t.Errorf("Unexpected failure to close")
	}
}
