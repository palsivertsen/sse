package sse

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type ResponseRecorderWrapper struct {
	*httptest.ResponseRecorder
	closer chan bool
}

func (r ResponseRecorderWrapper) CloseNotify() <-chan bool {
	return r.closer
}

func TestStream(t *testing.T) {
	var events = []struct {
		obj Event
		raw string
	}{
		{
			obj: Event{
				Type:    "test",
				Message: "this message\nhas two lines",
			},
			raw: "event: test\ndata: this message\ndata: has two lines\n\n",
		}, {
			obj: Event{Message: "No type defaults to 'message'"},
			raw: "event: message\ndata: No type defaults to 'message'\n\n",
		},
	}
	for _, event := range events {
		stream := NewStream()
		defer stream.Close()
		go stream.Send(event.obj)

		recorder := ResponseRecorderWrapper{
			ResponseRecorder: httptest.NewRecorder(),
			closer:           make(chan bool),
		}
		go func() {
			time.Sleep(time.Second * 2) // Wait for data
			recorder.closer <- true
		}()
		stream.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

		assert.Equal(t, "text/event-stream", recorder.Header().Get("content-type"))
		assert.Equal(t, event.raw, recorder.Body.String())
	}
}
