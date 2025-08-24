package request

import (
	"fmt"
	"io"
)

var ErrBadRequest = fmt.Errorf("bad request string")

const (
	bufferSize = 8
)

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	dataBuffer := make([]byte, 0)
	r := newRequest()
	err := parseRequestFromReader(reader, buf, dataBuffer, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func parseRequestFromReader(reader io.Reader, buf []byte, dataBuffer []byte, r *Request) error {
	for r.State != DoneState {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n > 0 {
			dataBuffer = append(dataBuffer, buf[:n]...)
		}
		consumed, parseErr := r.parse(dataBuffer)
		if parseErr != nil {
			return parseErr
		}
		if consumed > 0 {
			dataBuffer = dataBuffer[consumed:]
		}
		if err == io.EOF {
			break
		}
	}
	if r.State != DoneState {
		return ErrBadRequest
	}
	return nil
}
