package sse

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	testEvents := []struct {
		event    Event
		expected string
	}{
		{
			event:    Event{},
			expected: "",
		}, {
			event: Event{
				Comment: "This is a comment!",
			},
			expected: ":This is a comment!\n\n",
		}, {
			event: Event{
				Comment: `This is a
multiline comment!`,
			},
			expected: `:This is a
:multiline comment!

`,
		}, {
			event: Event{
				Message: "my message",
			},
			expected: "data:my message\n\n",
		}, {
			event: Event{
				Message: "my\nmultiline\nmessage",
			},
			expected: `data:my
data:multiline
data:message

`,
		}, {
			event: Event{
				Retry: &[]int{123}[0],
			},
			expected: "retry:123\n\n",
		}, {
			event: Event{
				Type: "new-event",
			},
			expected: "event:new-event\n\n",
		}, {
			event: Event{
				Retry:   &[]int{123}[0],
				Message: "my\nmessage",
				Comment: "with a\ncomment!",
				Type:    "new-event",
			},
			expected: `event:new-event
data:my
data:message
:with a
:comment!
retry:123

`,
		},
	}

	for _, testEvent := range testEvents {
		var buf bytes.Buffer
		testEvent.event.Encode(&buf)
		assert.Equal(t, testEvent.expected, buf.String())
	}
}
