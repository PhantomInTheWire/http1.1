package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"strconv"
)

type WriterState int

const (
	StateInitial WriterState = iota
	StateStatusWritten
	StateHeadersWritten
	StateBodyWritten
)

type Writer struct {
	w     io.Writer
	state WriterState
}

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

func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w, state: StateInitial}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != StateInitial {
		return fmt.Errorf("cannot write status line: already written or out of order")
	}
	if err := WriteStatusLine(w.w, statusCode); err != nil {
		return err
	}
	w.state = StateStatusWritten
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != StateStatusWritten {
		return fmt.Errorf("cannot write headers: status line not written yet")
	}
	if err := WriteHeaders(w.w, headers); err != nil {
		return err
	}
	w.state = StateHeadersWritten
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != StateHeadersWritten {
		return 0, fmt.Errorf("cannot write body: headers not written yet")
	}
	n, err := w.w.Write(p)
	if err != nil {
		return n, err
	}
	w.state = StateBodyWritten
	return n, nil
}

func (w *Writer) WriteChunk(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	_, err := fmt.Fprintf(w.w, "%x\r\n", len(p))
	if err != nil {
		return 0, err
	}

	n, err := w.w.Write(p)
	if err != nil {
		return n, err
	}

	_, err = w.w.Write([]byte("\r\n"))
	if err != nil {
		return n, err
	}

	return n, nil
}

func (w *Writer) WriteChunkedBodyDone() error {
	// Write the final zero-length chunk
	_, err := w.w.Write([]byte("0\r\n\r\n"))
	return err
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	w.w.Write(header)
}
