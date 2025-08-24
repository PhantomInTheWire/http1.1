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
