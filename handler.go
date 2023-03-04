package sse

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Event struct {
	Name  string
	ID    string
	Data  io.Reader
	Retry uint
}

func MarshalEvent(w io.Writer, e *Event) error {
	if e.Retry != 0 {
		fmt.Fprintf(w, "retry:%d\n", e.Retry) // TODO handle error
	}
	if e.Name != "" {
		writeField(w, "event", strings.NewReader(e.Name)) // TODO handle error
	}
	if e.ID != "" {
		writeField(w, "id", strings.NewReader(e.ID)) // TODO handle error
	}
	if e.Data != nil {
		scanner := bufio.NewScanner(e.Data)
		for scanner.Scan() {
			lineReader := strings.NewReader(scanner.Text())
			writeField(w, "data", lineReader) // TODO handle error
		}
	}

	// TODO scanner error

	if e.Name != "" || e.Data != nil || e.ID != "" || e.Retry != 0 {
		fmt.Fprintln(w) // TODO handle error
	}
	return nil
}

func writeField(w io.Writer, name string, content io.Reader) error {
	fmt.Fprintf(w, "%s:", name)
	io.Copy(w, content)
	fmt.Fprint(w, "\n")
	return nil
}
