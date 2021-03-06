package sse

import (
	"net/http"
	"sync"
)

// Stream is a HTTP handler for SSE
// Use NewStream to creat new instance
type Stream struct {
	events          chan Event
	closeNotifier   chan struct{}
	stop            chan struct{}
	serveHTTPMux    sync.Mutex
	serveHTTPCalled bool
}

// NewStream initializes a stream handler
func NewStream() *Stream {
	return &Stream{
		events:        make(chan Event),
		closeNotifier: make(chan struct{}),
		stop:          make(chan struct{}),
	}
}

// ServeHTTP sets up a stream (SSE) connection to the client
// Panics if called more than once
// Panics if http.ReposneWriter doesn't implement http.Flusher and http.CloseNotifier
// Stream is to be considered closed on return
func (s *Stream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Ensure only one call
	s.serveHTTPMux.Lock()
	if s.serveHTTPCalled {
		s.serveHTTPMux.Unlock()
		panic("ServeHTTP call only allowed once per Stream")
	}
	s.serveHTTPCalled = true
	s.serveHTTPMux.Unlock()
	flusher := w.(http.Flusher)
	clientCloseNotifier := w.(http.CloseNotifier)
	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()
	defer close(s.stop)
	for {
		select {
		case event := <-s.events:
			if event.Comment == "" && event.Type == "" {
				event.Type = "message"
			}
			event.Encode(w)
			flusher.Flush()
		case <-clientCloseNotifier.CloseNotify():
			return
		case <-s.closeNotifier:
			return
		}
	}
}

// Send an event to the stream
func (s *Stream) Send(event Event) {
	select {
	case <-s.stop:
	case s.events <- event:
	}
}

// Comment sends a comment to the stream
func (s *Stream) Comment(comment string) {
	s.Send(Event{
		Comment: comment,
	})
}

// Ping sends a ping comment to the stream
func (s *Stream) Ping() {
	s.Comment("ping")
}

// Retry tells the client how long in milliseconds to wait before trying to reconnect
func (s *Stream) Retry(milliseconds int) {
	s.Send(Event{Retry: &milliseconds})
}

// Close the connection to the client
// All functions becomes no-ops
func (s *Stream) Close() {
	select {
	case <-s.closeNotifier:
	default:
		close(s.closeNotifier)
	}
}

// CloseNotify returns a channel that will be closed when the stream is closed
func (s *Stream) CloseNotify() <-chan struct{} {
	return s.closeNotifier
}
