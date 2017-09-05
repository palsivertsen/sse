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
}

func (e *Event) Encode(w io.Writer) {
	if e.Type != "" {
		fmt.Fprint(w, "event: ")
		fmt.Fprintln(w, e.Type)
	}
	if e.Message != "" {
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
	}
	if e.Comment != "" {
		fmt.Fprint(w, ":")
		comment := bytes.NewBuffer([]byte(e.Message))
		for {
			b, err := comment.ReadByte()
			if err != nil {
				break
			}
			w.Write([]byte{b})
			if b == '\n' {
				fmt.Fprint(w, ": ")
			}
		}
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w)
}
