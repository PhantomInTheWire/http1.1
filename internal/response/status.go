package response

import (
	"fmt"
	"io"
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	var reason string
	switch statusCode {
	case 200:
		reason = "OK"
	case 400:
		reason = "Bad Request"
	case 404:
		reason = "Not Found"
	case 500:
		reason = "Internal Server Error"
	default:
		reason = "Status"
	}
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, reason)
	_, err := w.Write([]byte(statusLine))
	return err
}
