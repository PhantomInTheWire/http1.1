package request

import (
	"httpfromtcp/internal/headers"
)

type ParseState int

const (
	PendingState ParseState = iota
	ParsingHeadersState
	ParsingBodyState
	DoneState
)

type Request struct {
	RequestLine RequestLine
	State       ParseState
	Headers     headers.Headers
	Body        []byte
}

func newRequest() Request {
	return Request{
		State:   PendingState,
		Headers: headers.NewHeaders(),
	}
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}
