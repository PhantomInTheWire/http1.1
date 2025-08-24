package request

import (
	"strings"
)

func validateRequestLineFormat(requestLine string) error {
	parts := strings.Split(requestLine, spaceDelimiter)
	if len(parts) != expectedRequestLineParts {
		return ErrBadRequest
	}
	return nil
}

func validateHttpVersion(version string) bool {
	return version == supportedHttpVersion
}
