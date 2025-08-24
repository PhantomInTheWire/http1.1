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

	requestLine := parts[0]
	if err := validateRequestLineFormat(requestLine); err != nil {
		return nil, 0, err
	}

	rl, err := parseRequestLineComponents(requestLine)
	if err != nil {
		return nil, 0, err
	}

	return rl, len(requestLine) + len("\r\n"), nil
}

func parseRequestLineComponents(requestLine string) (*RequestLine, error) {
	parts := strings.Split(requestLine, " ")
	httpVersion, err := extractHttpVersion(parts[2])
	if err != nil {
		return nil, err
	}

	requestTarget := parts[1]
	method := parts[0]

	rl := &RequestLine{
		HttpVersion:   httpVersion,
		RequestTarget: requestTarget,
		Method:        method,
	}
	return rl, nil
}

func extractHttpVersion(versionPart string) (string, error) {
	parts := strings.Split(versionPart, "/")
	if len(parts) != 2 {
		return "", ErrBadRequest
	}
	httpVersion := parts[1]
	if !validateHttpVersion(httpVersion) {
		return "", fmt.Errorf("%s: %s", ErrBadRequest.Error(), versionPart)
	}
	return httpVersion, nil
}
