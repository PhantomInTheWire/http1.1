package headers

import (
	"fmt"
	"strings"
)

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	headerString := string(data)

	if strings.HasPrefix(headerString, "\r\n\r\n") {
		return len("\r\n\r\n"), true, nil
	}

	if strings.HasPrefix(headerString, "\r\n") {
		return len("\r\n"), true, nil
	}

	endIdx := strings.Index(headerString, "\r\n\r\n")
	if endIdx == -1 {
		return 0, false, nil
	}

	headerBlock := headerString[:endIdx]
	headerLines := strings.Split(headerBlock, "\r\n")

	for _, headerLine := range headerLines {
		if err := h.parseHeaderLine(headerLine); err != nil {
			return 0, false, err
		}
	}
	return endIdx + len("\r\n\r\n"), true, nil
}

func (h Headers) parseHeaderLine(headerLine string) error {
	if headerLine == "" {
		return nil
	}
	if hasSpaceBeforeColon(headerLine) {
		return fmt.Errorf("has a space before colon, header: %v", headerLine)
	}
	kv := strings.SplitN(headerLine, ":", 2)
	if len(kv) != 2 {
		return fmt.Errorf("malinformed header: %v lenkv: %v", headerLine, len(kv))
	}
	rawKey := strings.TrimSpace(kv[0])
	if rawKey == "" {
		return fmt.Errorf("empty header key")
	}
	if err := validateHeaderKey(rawKey); err != nil {
		return err
	}
	key := strings.ToLower(rawKey)
	value := strings.TrimSpace(kv[1])
	h.setHeaderValue(key, value)
	return nil
}

func (h Headers) setHeaderValue(key, value string) {
	if h[key] == "" {
		h[key] = value
	} else {
		h[key] += ", "
		h[key] += value
	}
}
