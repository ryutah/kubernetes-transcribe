package watch

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/ryutah/kubernetes-transcribe/pkg/api"
	"github.com/ryutah/kubernetes-transcribe/pkg/runtime"
	"github.com/ryutah/kubernetes-transcribe/pkg/watch"
)

// APIEventDecoder implements the watch.Decoder interface for io.ReadClosers that
// have contents which consist of a series of api.WatchEvent objects encoded via JSON.
type APIEventDecoder struct {
	stream  io.ReadCloser
	decoder *json.Decoder
}

// NewAPIEventDecoder creates an APIEventDecoder for the given stream.
func NewAPIEventDecoder(stream io.ReadCloser) *APIEventDecoder {
	return &APIEventDecoder{
		stream:  stream,
		decoder: json.NewDecoder(stream),
	}
}

// Decode block until it can return the next object in the stream. Returns an error
// if the stream is closed or an object can't be decoded.
func (d *APIEventDecoder) Decode() (action watch.EventType, object runtime.Object, err error) {
	var got api.WatchEvent
	err = d.decoder.Decode(&got)
	if err != nil {
		return action, nil, err
	}
	switch got.Type {
	case watch.Added, watch.Modified, watch.Deleted:
		return got.Type, got.Object.Object, err
	}
	return action, nil, fmt.Errorf("got invalid watch event type: %v", got.Type)
}

// Close closes the underlying stream.
func (d *APIEventDecoder) Close() {
	d.stream.Close()
}
