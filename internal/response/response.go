package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
)

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	statusLine := "HTTP/1.1 "
	switch statusCode {
	case StatusOK:
		statusLine += "200 OK\r\n"
	case StatusBadRequest:
		statusLine += "400 Bad Request\r\n"
	case StatusInternalServerError:
		statusLine += "500 Internal Server Error\r\n"
	default:
		return fmt.Errorf("invalid status code %v: ", statusCode)
	}
	_, err := w.Write([]byte(statusLine))
	if err != nil {
		return err
	}
	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()
	h["Content-Length"] = strconv.Itoa(contentLen)
	h["Connection"] = "close"
	h["Content-Type"] = "text/plain"
	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {
		_, err := fmt.Fprintf(w, "%v: %v\r\n", k, v)
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	return nil
}

func WriteBody(w io.Writer, body string) error {
	_, err := w.Write([]byte(body))
	if err != nil {
		return err
	}
	return nil
}
