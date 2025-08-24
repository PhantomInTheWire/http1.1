package request

import (
	"fmt"
	"strings"
)

func printRequest(r *Request) {
	printRequestLine(r)
	printHeaders(r)
}

func printRequestLine(r *Request) {
	fmt.Println("Request line:")
	fmt.Printf("- Method: %s\n", r.RequestLine.Method)
	fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
	fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)
}

func printHeaders(r *Request) {
	fmt.Println("Headers:")
	for key, value := range r.Headers {
		fmt.Printf("- %s: %s\n", key, value)
	}
}

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
