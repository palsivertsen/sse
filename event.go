package sse

import (
	"bufio"
	"errors"
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

type FieldContentError struct {
	FieldName string

	err error
}

func (e FieldContentError) Error() string {
	return fmt.Sprintf("field %q: %s", e.FieldName, e.err)
}

func (e FieldContentError) Unwrap() error { return e.err }

func WriteEvent(w io.Writer, e *Event) error {
	if e.Retry != 0 {
		_, err := fmt.Fprintf(w, "retry:%d\n", e.Retry)
		if err != nil {
			return fmt.Errorf("write: %w", err)
		}
	}
	if e.Name != "" {
		if err := writeField(w, "event", strings.NewReader(e.Name)); err != nil {
			return err
		}
	}
	if e.ID != "" {
		if err := writeField(w, "id", strings.NewReader(e.ID)); err != nil {
			return err
		}
	}
	if e.Data != nil {
		scanner := bufio.NewScanner(e.Data)
		for scanner.Scan() {
			lineReader := strings.NewReader(scanner.Text())
			if err := writeField(w, "data", lineReader); err != nil {
				return err
			}
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("scan: %w", err)
		}
	}

	if e.Name != "" || e.Data != nil || e.ID != "" || e.Retry != 0 {
		if _, err := fmt.Fprintln(w); err != nil {
			return fmt.Errorf("write end of event marker: %w", err)
		}
	}
	return nil
}

func writeField(w io.Writer, name string, content io.Reader) error {
	r := blacklistReader{
		r:     content,
		runes: []rune{'\n', '\r'},
	}
	if _, err := fmt.Fprintf(w, "%s:", name); err != nil {
		return FieldContentError{
			FieldName: name,
			err:       fmt.Errorf("write name: %w", err),
		}
	}
	if _, err := io.Copy(w, &r); err != nil {
		return FieldContentError{
			FieldName: name,
			err:       fmt.Errorf("write content: %w", err),
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return FieldContentError{
			FieldName: name,
			err:       fmt.Errorf("write end of line marker: %w", err),
		}
	}
	return nil
}

type blacklistReader struct {
	r     io.Reader
	runes []rune
}

func (r *blacklistReader) Read(bytes []byte) (int, error) {
	i, err := r.r.Read(bytes)
	if err != nil && !errors.Is(err, io.EOF) {
		return i, fmt.Errorf("read: %w", err)
	}

	s := string(bytes[:i])
	for _, rr := range r.runes {
		if strings.ContainsRune(s, rr) {
			return i, ErrMultilineContent
		}
	}
	return i, err
}
