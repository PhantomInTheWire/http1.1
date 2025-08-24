package headers

import (
	"fmt"
	"strings"
)

const (
	// Header key-value parsing
	headerKeyValuePartsCount = 2
	headerKeyIndex           = 0
	headerValueIndex         = 1

	// String literals
	emptyString            = ""
	colonDelimiter         = ":"
	headerValueSeparator   = ", "
	carriageReturnLineFeed = "\r\n"
	headersEndMarker       = "\r\n\r\n"

	// Character validation
	asciiSpace = 32
	asciiTilde = 126
	colonRune  = ':'
)

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	headerString := string(data)

	if strings.HasPrefix(headerString, headersEndMarker) {
		return len(headersEndMarker), true, nil
	}

	if strings.HasPrefix(headerString, carriageReturnLineFeed) {
		return len(carriageReturnLineFeed), true, nil
	}

	endIdx, found := findHeaderBlockEnd(headerString)
	if !found {
		return 0, false, nil
	}

	if err := h.processHeaderBlock(headerString[:endIdx]); err != nil {
		return 0, false, err
	}

	return endIdx + len(headersEndMarker), true, nil
}

func findHeaderBlockEnd(headerString string) (int, bool) {
	endIdx := strings.Index(headerString, headersEndMarker)
	return endIdx, endIdx != -1
}

func (h Headers) processHeaderBlock(headerBlock string) error {
	headerLines := strings.Split(headerBlock, carriageReturnLineFeed)
	for _, headerLine := range headerLines {
		if err := h.parseHeaderLine(headerLine); err != nil {
			return err
		}
	}
	return nil
}

func (h Headers) parseHeaderLine(headerLine string) error {
	if headerLine == emptyString {
		return nil
	}

	if err := validateHeaderLineFormat(headerLine); err != nil {
		return err
	}

	key, value, err := parseHeaderKeyValue(headerLine)
	if err != nil {
		return err
	}

	h.setHeaderValue(key, value)
	return nil
}

func validateHeaderLineFormat(headerLine string) error {
	if hasSpaceBeforeColon(headerLine) {
		return fmt.Errorf("has a space before colon, header: %v", headerLine)
	}
	return nil
}

func parseHeaderKeyValue(headerLine string) (string, string, error) {
	kv := strings.SplitN(headerLine, colonDelimiter, headerKeyValuePartsCount)
	if len(kv) != headerKeyValuePartsCount {
		return emptyString, emptyString, fmt.Errorf("malinformed header: %v lenkv: %v", headerLine, len(kv))
	}

	rawKey := strings.TrimSpace(kv[headerKeyIndex])
	if rawKey == emptyString {
		return emptyString, emptyString, fmt.Errorf("empty header key")
	}

	if err := validateHeaderKey(rawKey); err != nil {
		return emptyString, emptyString, err
	}

	key := strings.ToLower(rawKey)
	value := strings.TrimSpace(kv[headerValueIndex])
	return key, value, nil
}

func (h Headers) setHeaderValue(key, value string) {
	if h[key] == emptyString {
		h[key] = value
	} else {
		h[key] += headerValueSeparator
		h[key] += value
	}
}

func (h Headers) Get(key string) (string, error) {
	if err := validateHeaderKey(key); err != nil {
		return "", err
	}
	return h[strings.ToLower(key)], nil
}
