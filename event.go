package sse

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/palsivertsen/goutils/converters"
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
	e.lineWithTag("event", t)
}

func (e *eventWriter) writeMessage(message string) {
	e.linesWithTag("data", bytes.NewBufferString(message))
}

func (e *eventWriter) writeComment(c string) {
	e.linesWithTag("", bytes.NewBufferString(c))
}

func (e *eventWriter) writeRetry(retry int) {
	e.lineWithTag("retry", converters.Int(retry).ToString())
}

func (e *eventWriter) linesWithTag(tag string, r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		e.lineWithTag(tag, scanner.Text())
	}
}

func (e *eventWriter) lineWithTag(tag string, line string) {
	fmt.Fprintf(e, "%s:%s\n", tag, line)
}
