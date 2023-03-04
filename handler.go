package sse

import (
	"fmt"
	"io"
	"net/http"
)

type Handler interface {
	ServeSSE(*ResponseWriter, *http.Request)
}

type ResponseWriter struct {
	w io.Writer
}

func NewResponseWriter(opts ...func(*ResponseWriter)) *ResponseWriter {
	var rw ResponseWriter
	for _, opt := range opts {
		opt(&rw)
	}
	return &rw
}

func WithHTTPResponseWriter(httpRW http.ResponseWriter) func(*ResponseWriter) {
	return func(rw *ResponseWriter) {
		rw.w = httpRW
	}
}

func (rw *ResponseWriter) PushEvent(event *Event) error {
	if err := WriteEvent(rw.w, event); err != nil {
		return fmt.Errorf("write event: %w", err)
	}
	return nil
}

func NewHTTPHandler(handler Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "text/event-stream")
		handler.ServeSSE(
			NewResponseWriter(WithHTTPResponseWriter(rw)),
			r,
		)
	})
}
