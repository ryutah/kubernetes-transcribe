package watch

import (
	"reflect"
	"testing"
)

func TestFilter(t *testing.T) {
	table := []Event{
		{Added, testType("foo")},
		{Added, testType("bar")},
		{Added, testType("baz")},
		{Added, testType("qux")},
		{Added, testType("zoo")},
	}

	source := NewFake()
	filterd := Filter(source, func(e Event) (Event, bool) {
		return e, e.Object.(testType)[0] != 'b'
	})

	go func() {
		for _, item := range table {
			source.Action(item.Type, item.Object)
		}
		source.Stop()
	}()

	var got []string
	for {
		event, ok := <-filterd.ResultChan()
		if !ok {
			break
		}
		got = append(got, string(event.Object.(testType)))
	}
	if e, a := []string{"foo", "qux", "zoo"}, got; !reflect.DeepEqual(e, a) {
		t.Errorf("got %v, wanted %v", e, a)
	}
}

func TestFilterStop(t *testing.T) {
	source := NewFake()
	filtered := Filter(source, func(e Event) (Event, bool) {
		return e, e.Object.(testType)[0] != 'b'
	})

	go func() {
		source.Add(testType("foo"))
		filtered.Stop()
	}()

	var got []string
	for {
		event, ok := <-filtered.ResultChan()
		if !ok {
			break
		}
		got = append(got, string(event.Object.(testType)))
	}

	if e, a := []string{"foo"}, got; !reflect.DeepEqual(e, a) {
		t.Errorf("got %v, wanted %v", e, a)
	}
}
