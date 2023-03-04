package sse_test

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/palsivertsen/sse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		event *sse.Event
		out   string
	}{
		{event: &sse.Event{}},
		{
			event: &sse.Event{
				Data: strings.NewReader("This is the event data"),
			},
			out: "data:This is the event data\n\n",
		},
		{
			event: &sse.Event{
				Data: strings.NewReader("multiline\nevent"),
			},
			out: "data:multiline\ndata:event\n\n",
		},
		{
			event: &sse.Event{
				Name: "named event",
			},
			out: "event:named event\n\n",
		},
		{
			event: &sse.Event{
				ID: "this is the ID",
			},
			out: "id:this is the ID\n\n",
		},
		{
			event: &sse.Event{
				Retry: 1234,
			},
			out: "retry:1234\n\n",
		},
		{
			event: &sse.Event{
				Name:  "test",
				ID:    "cache key",
				Data:  strings.NewReader("this is a\nfull test event"),
				Retry: 888,
			},
			out: "retry:888\nevent:test\nid:cache key\ndata:this is a\ndata:full test event\n\n",
		},
	}
	for testNum, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("%d", testNum), func(t *testing.T) {
			t.Parallel()
			t.Logf("input:\n%s", spew.Sdump(tt))

			var buf bytes.Buffer
			err := sse.WriteEvent(&buf, tt.event)
			require.NoError(t, err)
			assert.Equal(t, tt.out, buf.String())
		})
	}
}

func TestWriteEvent_FieldFormatErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		event *sse.Event
		field string
		err   error
	}{
		{
			event: &sse.Event{
				Name: "multi\nline",
			},
			field: "event",
			err:   sse.ErrMultilineContent,
		},
		{
			event: &sse.Event{
				ID: "multi\nline",
			},
			field: "id",
			err:   sse.ErrMultilineContent,
		},
	}
	for testNum, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("%d", testNum), func(t *testing.T) {
			t.Parallel()
			t.Logf("input:\n%s", spew.Sdump(tt))

			err := sse.WriteEvent(io.Discard, tt.event)
			var fce sse.FieldContentError
			require.ErrorAs(t, err, &fce)
			assert.Equal(t, tt.field, fce.FieldName)
			assert.ErrorIs(t, err, tt.err)
		})
	}
}
