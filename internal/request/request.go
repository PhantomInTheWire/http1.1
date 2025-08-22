package request

import (
	"fmt"
	"io"
	"strings"
)

var BAD_REQUEST = fmt.Errorf("bad request string") 

const bufferSize = 8

func validateHttpVersion(version string) (bool) {
	return version == "1.1"
}

type ParseState int
const (
	PendingState ParseState = 0
	DoneState    ParseState = 1
)

type Request struct {
	RequestLine RequestLine
	State ParseState
}

func newRequest() Request {
	return Request{
		State: PendingState,
	}
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *Request) parse(data []byte) (int, error) {
	if r.State == DoneState {
			return 0, nil
		}
		
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
		r.State = DoneState
		return consumed, nil
}


func parseRequestLine(RequestStr string) (*RequestLine, int, error) {
	parts := strings.Split(RequestStr, "\r\n")
	if len(parts) == 1 && !strings.HasSuffix(RequestStr, "\r\n") {
		return nil, len(parts), nil
	}
	if len(parts) < 1 {
		return  nil, 0, BAD_REQUEST
	}
	
	seperatedLineOne := strings.Split(parts[0], " ")
	
	if len(seperatedLineOne) != 3 {
		return nil, 0, BAD_REQUEST
	}
	
	httpVersion := strings.Split(seperatedLineOne[2], "/")[1]
	requestTarget := seperatedLineOne[1]
	method := seperatedLineOne[0]
	if (!validateHttpVersion(httpVersion)) {
		return nil, 0, fmt.Errorf(BAD_REQUEST.Error(), RequestStr)
	}
	rl := &RequestLine{
		HttpVersion : httpVersion,
		RequestTarget : requestTarget,
		Method: method, 
	}
	return rl, len(parts[0]), nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	dataBuffer := make([]byte, 0)
	r := newRequest()
	for r.State != DoneState {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if n > 0 {
			dataBuffer = append(dataBuffer, buf[:n]...)
		}
		consumed, parseErr := r.parse(dataBuffer)
		if parseErr != nil {
			return nil, parseErr
		}
		if consumed > 0 {
			dataBuffer = dataBuffer[consumed:]
		}
		if err == io.EOF {
			break
		}
	}
	if r.State != DoneState {
			return nil, BAD_REQUEST
		}
		
	return &r, nil
}
