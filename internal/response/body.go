package response

import "io"

func WriteBody(w io.Writer, body string) error {
	_, err := w.Write([]byte(body))
	if err != nil {
		return err
	}
	return nil
}
