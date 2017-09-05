package sse

import (
	"bytes"
	"fmt"
	"io"
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
