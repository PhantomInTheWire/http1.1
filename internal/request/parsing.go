package request

import (
	"fmt"
	"strings"
)

func (r *Request) ParseRequestLine(data []byte) (int, error) {
	requestLine, consumed, err := parseRequestLine(string(data))
	if err != nil {
		return 0, err
	}
	if consumed == 0 {
		return 0, nil
	}
	if requestLine == nil {
		return 0, nil
	}

	r.RequestLine = *requestLine
	r.State = requestStateParsingHeaders
	return consumed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	if r.State == DoneState {
		return 0, nil
	}
	if r.State == PendingState {
		n, err := r.ParseRequestLine(data)
		if err != nil {
			return 0, err
		}
		return n, nil
	}
	if r.State == requestStateParsingHeaders {
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.State = DoneState
			return n, nil
		}
		return n, nil
	}
	return 0, ErrBadRequest
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.State != DoneState {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return totalBytesParsed, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func parseRequestLine(RequestStr string) (*RequestLine, int, error) {
	parts := strings.Split(RequestStr, "\r\n")
	if len(parts) == 1 && !strings.HasSuffix(RequestStr, "\r\n") {
		return nil, 0, nil
	}
	if len(parts) < 1 {
		return nil, 0, ErrBadRequest
	}

	seperatedLineOne := strings.Split(parts[0], " ")

	if len(seperatedLineOne) != 3 {
		return nil, 0, ErrBadRequest
	}

	httpVersion := strings.Split(seperatedLineOne[2], "/")[1]
	requestTarget := seperatedLineOne[1]
	method := seperatedLineOne[0]
	if !validateHttpVersion(httpVersion) {
		return nil, 0, fmt.Errorf("%s: %s", ErrBadRequest.Error(), RequestStr)
	}
	rl := &RequestLine{
		HttpVersion:   httpVersion,
		RequestTarget: requestTarget,
		Method:        method,
	}
	return rl, len(parts[0]) + len("\r\n"), nil
}
