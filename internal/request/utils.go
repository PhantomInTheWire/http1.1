package request

import "strings"

// Constants for HTTP request validation
const (
	supportedHttpVersion = "1.1"
)

func validateHttpVersion(version string) bool {
	return version == supportedHttpVersion
}

func validateRequestLineFormat(requestLine string) error {
	parts := strings.Split(requestLine, spaceDelimiter)
	if len(parts) != expectedRequestLineParts {
		return ErrBadRequest
	}
	return nil
}
