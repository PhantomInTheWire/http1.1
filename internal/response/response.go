package response

import (
	"httpfromtcp/internal/headers"
	"io"
)

func WriteSimpleResponse(w io.Writer, statusCode StatusCode, contentType string, body string) error {
	writer := NewWriter(w)

	if err := writer.WriteStatusLine(statusCode); err != nil {
		return err
	}

	h := GetDefaultHeaders(len(body))
	if contentType != "" {
		h["Content-Type"] = contentType
	}
	if err := writer.WriteHeaders(h); err != nil {
		return err
	}

	if _, err := writer.WriteBody([]byte(body)); err != nil {
		return err
	}

	return nil
}

func WriteChunkedResponse(w io.Writer, statusCode StatusCode, contentType string, chunks [][]byte, trailers headers.Headers) error {
	writer := NewWriter(w)

	if err := writer.WriteStatusLine(statusCode); err != nil {
		return err
	}

	h := headers.NewHeaders()
	h["Transfer-Encoding"] = "chunked"
	if contentType != "" {
		h["Content-Type"] = contentType
	}
	h["Connection"] = "close"
	if err := writer.WriteHeaders(h); err != nil {
		return err
	}

	for _, chunk := range chunks {
		if _, err := writer.WriteChunk(chunk); err != nil {
			return err
		}
	}

	if err := writer.WriteChunkedBodyDone(); err != nil {
		return err
	}

	if len(trailers) > 0 {
		if err := writer.WriteTrailers(trailers); err != nil {
			return err
		}
	}

	return nil
}

func WriteErrorResponse(w io.Writer, statusCode StatusCode, message string) error {
	return WriteSimpleResponse(w, statusCode, "text/plain", message)
}

func WriteJSONResponse(w io.Writer, statusCode StatusCode, jsonBody string) error {
	return WriteSimpleResponse(w, statusCode, "application/json", jsonBody)
}
