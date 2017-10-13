package sse

import (
	"net/http"
)

type Stream struct {
	events        chan Event
	closeNotifier chan bool
}

func NewStream() *Stream {
	return &Stream{
		events:        make(chan Event),
		closeNotifier: make(chan bool, 1),
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
		}
	}
}

func (s *Stream) Send(event Event) {
	s.events <- event
}

func (s *Stream) Comment(comment string) {
	s.Send(Event{
		Comment: comment,
	})
}

func (s *Stream) Ping() {
	s.Comment("ping")
}

func (s *Stream) Retry(retry int) {
	s.Send(Event{Retry: &retry})
}

func (s *Stream) Close() {
	close(s.events)
	close(s.closeNotifier)
}

func (s *Stream) CloseNotify() <-chan bool {
	return s.closeNotifier
}
