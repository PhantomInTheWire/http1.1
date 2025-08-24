package request

import "strings"

func validateHttpVersion(version string) bool {
	return version == "1.1"
}

func validateRequestLineFormat(requestLine string) error {
	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		return ErrBadRequest
	}
	return nil
}
