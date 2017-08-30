package sse

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type Event struct {
	Message string
	Type    string
}

func (e *Event) Encode(w io.Writer) {
	fmt.Fprint(w, "event: ")
	if e.Type == "" {
		fmt.Fprintln(w, "message")
	} else {
		fmt.Fprintln(w, e.Type)
	}
	fmt.Fprint(w, "data: ")
	data := bytes.NewBuffer([]byte(e.Message))
	for {
		b, err := data.ReadByte()
		if err != nil {
			break
		}
		w.Write([]byte{b})
		if b == '\n' {
			fmt.Fprint(w, "data: ")
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w)
}

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
