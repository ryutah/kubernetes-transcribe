package apiserver

import (
	"net/http"
	"time"

	"github.com/ryutah/kubernetes-transcribe/pkg/api"
	"github.com/ryutah/kubernetes-transcribe/pkg/httplog"
	"github.com/ryutah/kubernetes-transcribe/pkg/labels"
	"github.com/ryutah/kubernetes-transcribe/pkg/runtime"
)

type RESTHandler struct {
	storage     map[string]RESTStorage
	codec       runtime.Codec
	ops         *Operations
	asyncOpWait time.Duration
}

// ServeHTTP handles requests to all RESTStorage objects.
func (r *RESTHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	parts := splitPath(req.URL.Path)
	if len(parts) < 1 {
		notFound(w, req)
		return
	}
	storage := r.storage[parts[0]]
	if storage == nil {
		httplog.LogOf(req, w).Addf("'%v' has no storage obejct", parts[0])
		notFound(w, req)
		return
	}

	// XXX
}

// handleRESTStorage is the main dispatcher for a storage object.  It switches on the HTTP method, and then
// on path length, according to the following table:
//   Method     Path          Action
//   GET        /foo          list
//   GET        /foo/bar      get 'bar'
//   POST       /foo          create
//   PUT        /foo/bar      update 'bar'
//   DELETE     /foo/bar      delete 'bar'
// Returns 404 if the method/pattern doesn't match one of these entries
// The s accepts several query parameters:
//    sync=[false|true] Synchronous request (only applies to create, update, delete operations)
//    timeout=<duration> Timeout for Synchronous requests, only applies if sync=true
//    labels=<label-selector> Used for filtering list operations
func (r *RESTHandler) handleRESTStorage(parts []string, req *http.Request, w http.ResponseWriter, storage RESTStorage) {
	sync := req.URL.Query().Get("sync") == "true"
	timeout := parseTimeout(req.URL.Query().Get("timeout"))
	switch req.Method {
	case "GET":
		switch len(parts) {
		case 1:
			label, err := labels.ParseSelector(req.URL.Query().Get("labels"))
			if err != nil {
				errorJSON(err, r.codec, w)
				return
			}
			field, err := labels.ParseSelector(req.URL.Query().Get("fields"))
			if err != nil {
				errorJSON(err, r.codec, w)
				return
			}
			list, err := storage.List(label, field)
			if err != nil {
				errorJSON(err, r.codec, w)
				return
			}
			writeJSON(http.StatusOK, r.codec, list, w)
		case 2:
			item, err := storage.Get(parts[1])
			if err != nil {
				errorJSON(err, r.codec, w)
				return
			}
			writeJSON(http.StatusOK, r.codec, item, w)
		default:
			notFound(w, req)
		}

	case "POST":
		if len(parts) != 1 {
			notFound(w, req)
			return
		}
		body, err := readBody(req)
		if err != nil {
			errorJSON(err, r.codec, w)
			return
		}
		obj := storage.New()
		err = r.codec.DecodeInto(body, obj)
		if err != nil {
			errorJSON(err, r.codec, w)
			return
		}
		out, err := storage.Create(obj)
		if err != nil {
			errorJSON(err, r.codec, w)
			return
		}
		op := r.createOperation(out, sync, timeout)
		r.finishReq(op, req, w)

	case "DELETE":
		if len(parts) != 2 {
			notFound(w, req)
			return
		}
		out, err := storage.Delete(parts[1])
		if err != nil {
			errorJSON(err, r.codec, w)
			return
		}
		op := r.createOperation(out, sync, timeout)
		r.finishReq(op, req, w)

	case "PUT":
		if len(parts) != 2 {
			notFound(w, req)
			return
		}
		body, err := readBody(req)
		if err != nil {
			errorJSON(err, r.codec, w)
			return
		}
		obj := storage.New()
		err = r.codec.DecodeInto(body, obj)
		if err != nil {
			errorJSON(err, r.codec, w)
			return
		}
		out, err := storage.Update(obj)
		if err != nil {
			errorJSON(err, r.codec, w)
			return
		}
		op := r.createOperation(out, sync, timeout)
		r.finishReq(op, req, w)

	default:
		notFound(w, req)
	}
}

// createOperation creates an operation to process a channel response.
func (r *RESTHandler) createOperation(out <-chan runtime.Object, sync bool, timeout time.Duration) *Operation {
	op := r.ops.NewOperation(out)
	if sync {
		op.WaitFor(timeout)
	} else if r.asyncOpWait != 0 {
		op.WaitFor(r.asyncOpWait)
	}
	return op
}

// finishReq finishes up a request, waiting until the operation finishes or, after a timeout, creating an
// Operation to receive the result and returning its ID down the writer.
func (r *RESTHandler) finishReq(op *Operation, req *http.Request, w http.ResponseWriter) {
	obj, complete := op.StatusOrResult()
	if complete {
		status := http.StatusOK
		switch stat := obj.(type) {
		case *api.Status:
			if stat.Code != 0 {
				status = stat.Code
			}
		}
		writeJSON(status, r.codec, obj, w)
	} else {
		writeJSON(http.StatusAccepted, r.codec, obj, w)
	}
}
