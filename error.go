package sse

const (
	ErrMultilineContent = Error("multiline content not allowed")
)

type Error string

func (e Error) Error() string {
	return string(e)
}
