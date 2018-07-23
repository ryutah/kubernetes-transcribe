package api

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/ryutah/kubernetes-transcribe/pkg/runtime"
	"github.com/ryutah/kubernetes-transcribe/pkg/watch"
)

// WatchEvent objects are streamed from the api server in response to a watch request.
// These are not API objects and are unversioned today.
type WatchEvent struct {
	// The type of the event; added, modified, or deleted.
	Type watch.EventType

	// For added or modified objects, this is the new object; for deleted objects,
	// it's the state of the object immediately prior to its deletion.
	Object EmbeddedObject
}

// watchSerialization defines the JSON wire equivalent of watch.Evnet
type watchSerialization struct {
	Type   watch.EventType
	Object json.RawMessage
}

// NewJSONWatchEvnet returns an object that will serialize to JSON and back
// to WatchEvent
func NewJSONWatchEvnet(codec runtime.Codec, event watch.Event) (interface{}, error) {
	obj, ok := event.Object.(runtime.Object)
	if !ok {
		return nil, fmt.Errorf("The event object cannot be safely converted to JSON: %v", reflect.TypeOf(event.Object).Name())
	}
	data, err := codec.Encode(obj)
	if err != nil {
		return nil, err
	}
	return &watchSerialization{event.Type, json.RawMessage(data)}, nil
}
