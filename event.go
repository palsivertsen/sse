package sse

import (
	"bytes"
	"fmt"
	"io"
)

type Event struct {
	Message string
	Type    string
	Comment string
	Retry   *int
}

func (e *Event) Encode(w io.Writer) {
	wrapper := &eventWriter{writer: w}
	defer wrapper.end()

	if e.Type != "" {
		wrapper.writeType(e.Type)
	}
	if e.Message != "" {
		wrapper.writeMessage(e.Message)
	}
	if e.Comment != "" {
		wrapper.writeComment(e.Comment)
	}
	if e.Retry != nil {
		wrapper.writeRetry(*e.Retry)
	}
}

type eventWriter struct {
	writer  io.Writer
	written bool
}

func (e *eventWriter) Write(buf []byte) (n int, err error) {
	if !e.written {
		e.written = true
	}
	return e.writer.Write(buf)
}

func (e *eventWriter) end() {
	if e.written {
		fmt.Fprintln(e.writer)
	}
}

func (e *eventWriter) writeType(t string) {
	fmt.Fprintf(e, "event:%s\n", t)
}

func (e *eventWriter) writeMessage(message string) {
	fmt.Fprint(e, "data:")
	data := bytes.NewBuffer([]byte(message))
	for {
		b, err := data.ReadByte()
		if err != nil {
			break
		}
		e.Write([]byte{b})
		if b == '\n' {
			fmt.Fprint(e, "data:")
		}
	}
	fmt.Fprintln(e)
}

func (e *eventWriter) writeComment(c string) {
	fmt.Fprint(e, ":")
	comment := bytes.NewBufferString(c)
	for {
		b, err := comment.ReadByte()
		if err != nil {
			break
		}
		e.Write([]byte{b})
		if b == '\n' {
			fmt.Fprint(e, ":")
		}
	}
	fmt.Fprintln(e)
}

func (e *eventWriter) writeRetry(retry int) {
	fmt.Fprintf(e, "retry:%d\n", retry)
}
