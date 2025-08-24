package request

import (
	"fmt"
	"strings"
)

const (
	supportedHttpVersion = "1.1"
)

const (
	requestLineMethodIndex   = 0
	requestLineTargetIndex   = 1
	requestLineVersionIndex  = 2
	expectedRequestLineParts = 3
)

const (
	httpVersionPartsCount = 2
	httpVersionValueIndex = 1
)

const (
	requestLineDataIndex = 0
	minDataPartsCount    = 1
)

const (
	emptyString            = ""
	spaceDelimiter         = " "
	carriageReturnLineFeed = "\r\n"
	slashDelimiter         = "/"
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
