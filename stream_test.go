package sse

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			raw: "event:test\ndata:this message\ndata:has two lines\n\n",
		}, {
			obj: Event{Message: "No type defaults to 'message'"},
			raw: "event:message\ndata:No type defaults to 'message'\n\n",
		},
	}
	for _, event := range events {
		stream := NewStream()
		recorder := ResponseRecorderWrapper{
			ResponseRecorder: httptest.NewRecorder(),
			closer:           make(chan bool),
		}
		go stream.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
		go func() {
			time.Sleep(time.Second * 2) // Wait for data
			recorder.closer <- true
		}()
		stream.Send(event.obj)

		assert.Equal(t, "text/event-stream", recorder.Result().Header.Get("content-type"))
		assert.Equal(t, event.raw, recorder.Body.String())
	}
}

func TestStream_Flushed(t *testing.T) {
	stream := NewStream()
	recorder := ResponseRecorderWrapper{
		ResponseRecorder: httptest.NewRecorder(),
		closer:           make(chan bool),
	}
	go func() {
		time.Sleep(time.Millisecond * 500) // Wait for data
		recorder.closer <- true
	}()
	stream.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.True(t, recorder.Flushed, "Stream should be flushed on initial connection")
	assert.Equal(t, "text/event-stream", recorder.Result().Header.Get("content-type"))
	assert.Equal(t, http.StatusOK, recorder.Result().StatusCode)
}

func TestStream_Ping(t *testing.T) {
	stream := NewStream()
	recorder := ResponseRecorderWrapper{
		ResponseRecorder: httptest.NewRecorder(),
		closer:           make(chan bool),
	}
	go func() {
		stream.Ping()
		time.Sleep(time.Millisecond * 500) // Wait for data
		recorder.closer <- true
	}()
	stream.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, ":ping\n\n", recorder.Body.String())
}

func TestStream_Close(t *testing.T) {
	recorder := ResponseRecorderWrapper{
		ResponseRecorder: httptest.NewRecorder(),
		closer:           make(chan bool),
	}
	unit := NewStream()
	responseNotifier := make(chan struct{})
	go func() {
		unit.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
		responseNotifier <- struct{}{}
	}()

	unit.Ping()
	unit.Close()
	go require.NotPanics(t, unit.Ping, "Should not panic after close")

	select {
	case <-time.NewTimer(time.Second * 10).C:
		t.Fatal("Stream did not close connection")
	case <-responseNotifier:
	}
}

func TestStream_SendCloseDeadlock(t *testing.T) {
	recorder := ResponseRecorderWrapper{
		ResponseRecorder: httptest.NewRecorder(),
		closer:           make(chan bool),
	}
	unit := NewStream()
	go unit.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	unit.Ping()
	unit.Close()
	unit.Ping()
	unit.Ping()
	unit.Ping()
}

func TestStream_MultiClose(t *testing.T) {
	unit := NewStream()
	require.NotPanics(t, unit.Close)
	require.NotPanics(t, unit.Close)
}

func TestStream_MultiServeHTTP(t *testing.T) {
	recorder := ResponseRecorderWrapper{
		ResponseRecorder: httptest.NewRecorder(),
		closer:           make(chan bool),
	}
	unit := NewStream()

	go require.NotPanics(t, func() { unit.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil)) })
	time.Sleep(time.Second) // Give go routine time to start
	require.Panics(t, func() { unit.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil)) },
		"Second call to ServeHTTP should panic!")
	require.NotPanics(t, unit.Close)
}
