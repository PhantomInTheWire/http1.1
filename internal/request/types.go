package request

import "httpfromtcp/internal/headers"

type ParseState int

const (
	PendingState               ParseState = 0
	DoneState                  ParseState = 1
	requestStateParsingHeaders ParseState = 2
)

type Request struct {
	RequestLine RequestLine
	State       ParseState
	Headers     headers.Headers
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
