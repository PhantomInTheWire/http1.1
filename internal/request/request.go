package request

import (
	"fmt"
	"strings"
	"io"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

var BAD_REQUEST = fmt.Errorf("bad request string") 

func validateHttpVersion(version string) (bool) {
	return version == "1.1"
}

func parseRequestLine(RequestStr string) (*RequestLine, string, error) {
	parts := strings.Split(RequestStr, "\r\n")
	if len(parts) < 1 {
		return  nil, RequestStr, BAD_REQUEST
	}
	
	seperatedLineOne := strings.Split(parts[0], " ")
	
	if len(seperatedLineOne) != 3 {
		return nil, RequestStr, BAD_REQUEST
	}
	
	httpVersion := strings.Split(seperatedLineOne[2], "/")[1]
	requestTarget := seperatedLineOne[1]
	method := seperatedLineOne[0]
	if (!validateHttpVersion(httpVersion)) {
		return nil, RequestStr, fmt.Errorf(BAD_REQUEST.Error(), RequestStr)
	}
	rl := &RequestLine{
		HttpVersion : httpVersion,
		RequestTarget : requestTarget,
		Method: method, 
	}
	return rl, parts[2], nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	requestBytes, err := io.ReadAll(reader)
	if (err != nil) {
		return nil, err
	}
	requestLine, _, err := parseRequestLine(string(requestBytes))
	if (err != nil) {
		return nil, err
	}
	return &Request{RequestLine: *requestLine}, nil
}

