package httplog

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/golang/glog"
)

// Handler wraps all HTTP calls to delegate with nice logging.
// delegate may use LogOf(w).Addf(...) to write additional info to
// the per-request log message.
//
// Intended to wrap calls to your ServeMux.
func Handler(delegate http.Handler, pred StackTracePred) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer NewLogged(r, &w).StacktraceWhen(pred).Log()
		delegate.ServeHTTP(w, r)
	})
}

// StackTracePred returns true if a stacktrace should be logged for this status.
type StackTracePred func(httpStatus int) (logStackTrace bool)

type logger interface {
	Addf(format string, data ...interface{})
}

// Add a layer on top of ResponseWriter, so we can track latency and error
// message source.
type respLogger struct {
	status      int
	statusStack string
	addedInfo   string
	startTime   time.Time

	req *http.Request
	w   http.ResponseWriter

	logStackTracePred StackTracePred
}

// Simple logger that logs immediately when Addf is called
type passthroughLogger struct{}

func (passthroughLogger) Addf(format string, data ...interface{}) {
	glog.Infof(format, data...)
}

// DefaultStacktracePred is the default implementation of StackTracePred.
func DefaultStacktracePred(status int) bool {
	return status < http.StatusOK || status >= http.StatusBadRequest
}

// NewLogged turns a normal response writer into a logged response writer.
//
// Usage:
//
// defer NewLogged(req, &w).StackTraceWhen(StatusIsNot(200, 202)).Log()
//
// (Only the call to Log() is defered, so you can set everything up in one line!)
//
// Note that this *changes* your writer, to route response writing actions
// through the logger.
//
// Use LogOf(w).Addf(...) to log something along with the response result.
func NewLogged(req *http.Request, w *http.ResponseWriter) *respLogger {
	if _, ok := (*w).(*respLogger); ok {
		// Don't double-wrap
		panic("multiple NewLogged calls!")
	}
	rl := &respLogger{
		startTime:         time.Now(),
		req:               req,
		w:                 *w,
		logStackTracePred: DefaultStacktracePred,
	}
	*w = rl // hijack caller's writer!
	return rl
}

// LogOf returns the logger hiding in w. If there is not an existing logger
// then a passthroughLogger will be created which will log to stdout immediately
// when Addf is called.
func LogOf(req *http.Request, w http.ResponseWriter) logger {
	if _, exists := w.(*respLogger); !exists {
		pl := &passthroughLogger{}
		return pl
	}
	if rl, ok := w.(*respLogger); ok {
		return rl
	}
	panic("Unable to find or create the logger!")
}

// Unlogged returns the original ResponseWriter, or w if it is not our inserted logger.
func Unlogged(w http.ResponseWriter) http.ResponseWriter {
	if rl, ok := w.(*respLogger); ok {
		return rl.w
	}
	return w
}

// StacktraceWhen sets the stacktrace logging predicate, which decides when to log a stacktrace.
// There's a default. so you don't need to call this unless you don't like the default.
func (r *respLogger) StacktraceWhen(pred StackTracePred) *respLogger {
	r.logStackTracePred = pred
	return r
}

// StatusIsNot returns a StackTracePred which will cause stacktraces to be logged
// for any status *not* in the given list.
func StatusIsNot(statuses ...int) StackTracePred {
	return func(status int) bool {
		for _, s := range statuses {
			if status == s {
				return false
			}
		}
		return true
	}
}

// Addf adds additional data to be logged with this request.
func (r *respLogger) Addf(format string, data ...interface{}) {
	r.addedInfo += "\n" + fmt.Sprintf(format, data...)
}

func (r *respLogger) Log() {
	latency := time.Since(r.startTime)
	glog.Infof("%s %s: (%v) %v%v%v", r.req.Method, r.req.RequestURI, latency, r.status, r.statusStack, r.addedInfo)
}

// Header implements http.ResponseWriter.
func (r *respLogger) Header() http.Header {
	return r.w.Header()
}

// Write implements http.ResponseWriter.
func (r *respLogger) Write(b []byte) (int, error) {
	return r.w.Write(b)
}

// WriteHeader implements http.ResponseWriter.
func (r *respLogger) WriteHeader(status int) {
	r.status = status
	if r.logStackTracePred(status) {
		// Only log stacks for errors
		stack := make([]byte, 2048)
		stack = stack[:runtime.Stack(stack, false)]
		r.statusStack = "\n" + string(stack)
	} else {
		r.statusStack = ""
	}
	r.w.WriteHeader(status)
}
