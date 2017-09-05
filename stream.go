package sse

import (
	"net/http"
)

type Stream struct {
	events chan Event
	closed bool
}

func NewStream() *Stream {
	return &Stream{
		events: make(chan Event),
	}
}

func (s *Stream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher := w.(http.Flusher)
	closeNotifier := w.(http.CloseNotifier)
	w.Header().Set("Content-Type", "text/event-stream")
	for !s.closed {
		select {
		case event := <-s.events:
			event.Encode(w)
			flusher.Flush()
		case <-closeNotifier.CloseNotify():
			s.closed = true
		}
	}
}

func (s *Stream) Send(event Event) {
	s.events <- event
}

func (s *Stream) Close() {
	close(s.events)
	s.closed = true
}

func (s *Stream) Closed() bool {
	return s.closed
}
