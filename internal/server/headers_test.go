package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type flushRecorder struct {
	httptest.ResponseRecorder
	flushed bool
}

func (f *flushRecorder) Flush() {
	f.flushed = true
}

func TestResponseWriterDelegatesFlush(t *testing.T) {
	t.Parallel()
	rec := &flushRecorder{}
	rw := &responseWriter{ResponseWriter: rec, guard: &Guard{}}
	if _, ok := any(rw).(http.Flusher); !ok {
		t.Fatal("responseWriter should implement http.Flusher")
	}
	rw.Flush()
	if !rec.flushed {
		t.Fatal("expected underlying Flush")
	}
}
