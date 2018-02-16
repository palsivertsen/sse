package sse

import (
	"net/http"
)

// Stream is a HTTP handler for SSE
// Use NewStream to creat new instance
type Stream struct {
	events        chan Event
	closeNotifier chan bool
	close         chan struct{}
}

// NewStream initializes a stream handler
func NewStream() *Stream {
	return &Stream{
		events:        make(chan Event),
		closeNotifier: make(chan bool, 1),
		close:         make(chan struct{}),
	}
}

func (s *Stream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher := w.(http.Flusher)
	closeNotifier := w.(http.CloseNotifier)
	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()
	for {
		select {
		case event := <-s.events:
			if event.Comment == "" && event.Type == "" {
				event.Type = "message"
			}
			event.Encode(w)
			flusher.Flush()
		case <-closeNotifier.CloseNotify():
			s.closeNotifier <- true
			return
		case <-s.close:
			return
		}
	}
}

// Send an event to the stream
func (s *Stream) Send(event Event) {
	s.events <- event
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
func (s *Stream) Retry(retry int) {
	s.Send(Event{Retry: &retry})
}

// Close the connection to the client
// Avoid using the stream after it's closed
func (s *Stream) Close() {
	s.close <- struct{}{}
	close(s.closeNotifier)
}

// CloseNotify notifies if the connection to client was closed
// Avoid using the stream after client connection was closed
func (s *Stream) CloseNotify() <-chan bool {
	return s.closeNotifier
}
