package apiserver

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/websocket"

	"github.com/ryutah/kubernetes-transcribe/pkg/api"
	"github.com/ryutah/kubernetes-transcribe/pkg/httplog"
	"github.com/ryutah/kubernetes-transcribe/pkg/labels"
	"github.com/ryutah/kubernetes-transcribe/pkg/runtime"
	"github.com/ryutah/kubernetes-transcribe/pkg/watch"
)

type WatchHandler struct {
	storage map[string]RESTStorage
	codec   runtime.Codec
}

func getWatchParams(query url.Values) (label, field labels.Selector, resourceVersion uint64) {
	if s, err := labels.ParseSelector(query.Get("labels")); err != nil {
		label = labels.Everything()
	} else {
		label = s
	}
	if s, err := labels.ParseSelector(query.Get("fields")); err != nil {
		field = labels.Everything()
	} else {
		field = s
	}
	if rv, err := strconv.ParseUint(query.Get("resourceVersion"), 10, 64); err == nil {
		resourceVersion = rv
	}
	return label, field, resourceVersion
}

var connectionUpgradeRegex = regexp.MustCompile("(^|.*,\\s*)upgrade($|\\s*,)")

func isWebSocketRequest(req *http.Request) bool {
	return connectionUpgradeRegex.MatchString(strings.ToLower(req.Header.Get("Connection"))) && strings.ToLower(req.Header.Get("Upgrade")) == "websocket"
}

// ServeHTTP processes watch requests.
func (h *WatchHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	parts := splitPath(req.URL.Path)
	if len(parts) < 1 || req.Method != "GET" {
		notFound(w, req)
		return
	}
	storage := h.storage[parts[0]]
	if storage == nil {
		notFound(w, req)
		return
	}
	if watcher, ok := storage.(ResourceWatcher); ok {
		label, field, resourceVersion := getWatchParams(req.URL.Query())
		watching, err := watcher.Watch(label, field, resourceVersion)
		if err != nil {
			errorJSON(err, h.codec, w)
			return
		}
		// TODO: This is one watch per connection. We want to multiplex, so that
		// multiple watches of the same thing don't create two watches downstream.
		watchServer := &WatchServer{watching: watching, codec: h.codec}
		if isWebSocketRequest(req) {
			websocket.Handler(watchServer.HandleWS).ServeHTTP(httplog.Unlogged(w), req)
		} else {
			watchServer.ServeHTTP(w, req)
		}
		return
	}

	notFound(w, req)
}

// WatchServer serves a watch.Interface over a websocket or vanilla HTTP.
type WatchServer struct {
	watching watch.Interface
	codec    runtime.Codec
}

// HandleWS implements a websocket handler.
func (s *WatchServer) HandleWS(ws *websocket.Conn) {
	done := make(chan struct{})
	go func() {
		var unused interface{}
		// Expect this to block until the connection is closed, Client should not
		// send anything.
		websocket.JSON.Receive(ws, &unused)
		close(done)
	}()
	for {
		select {
		case <-done:
			s.watching.Stop()
			return
		case event, ok := <-s.watching.ResultChan():
			if !ok {
				// End of results.
				return
			}
			obj, err := api.NewJSONWatchEvnet(s.codec, event)
			if err != nil {
				// Client disconnect.
				s.watching.Stop()
				return
			}
			if err := websocket.JSON.Send(ws, obj); err != nil {
				// Client disconnect.
				s.watching.Stop()
				return
			}
		}
	}
}

// ServeHTTP serves a series of JSON encoded events via straight HTTP with
// Transfer-Encoding: chunked.
func (s *WatchServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	loggedW := httplog.LogOf(req, w)
	w = httplog.Unlogged(w)

	cn, ok := w.(http.CloseNotifier)
	if !ok {
		loggedW.Addf("unable to get CloseNotifier")
		http.NotFound(w, req)
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		loggedW.Addf("unable to get Flusher")
		http.NotFound(w, req)
		return
	}

	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	encoder := json.NewEncoder(w)
	for {
		select {
		case <-cn.CloseNotify():
			s.watching.Stop()
			return
		case event, ok := <-s.watching.ResultChan():
			if !ok {
				// End of results.
				return
			}
			obj, err := api.NewJSONWatchEvnet(s.codec, event)
			if err != nil {
				// Client disconnect.
				s.watching.Stop()
				return
			}
			if err := encoder.Encode(obj); err != nil {
				s.watching.Stop()
				return
			}
			flusher.Flush()
		}
	}
}
