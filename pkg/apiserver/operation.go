package apiserver

import (
	"net/http"
	"sync"
	"time"

	"github.com/ryutah/kubernetes-transcribe/pkg/runtime"
)

type OperationHandler struct {
	ops   *Operations
	codec runtime.Codec
}

func (o *OperationHandler) ServeHTTP(http.ResponseWriter, *http.Request) {
	panic("not implemented")
}

// Operation represents an ongoing action which the server is performing.
type Operation struct {
	ID       string
	result   runtime.Object
	awaiting <-chan runtime.Object
	finished *time.Time
	lock     sync.Mutex
	noftify  chan struct{}
}

// Operations tracks all the ongoing operations.
type Operations struct {
	// Access only using functions from atomic.
	lastID int64

	// 'lock' guards the ops map.
	lock sync.Mutex
	ops  map[string]*Operation
}

func NewOperations() *Operations {
	panic("")
}
