package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

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
	_, err := w.w.Write([]byte("0\r\n\r\n"))
	if err != nil {
		return err
	}
	w.state = StateBodyWritten
	return nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.state != StateBodyWritten {
		return fmt.Errorf("cannot write trailers: body not written yet")
	}
	for k, v := range h {
		_, err := fmt.Fprintf(w.w, "%v: %v\r\n", k, v)
		if err != nil {
			return err
		}
	}
	_, err := w.w.Write([]byte("\r\n"))
	return err
}
