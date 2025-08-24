package headers

import (
	"fmt"
	"strings"
)

func hasSpaceBeforeColon(s string) bool {
	idx := strings.IndexRune(s, ':')
	if idx == -1 || idx == 0 {
		return false
	}
	if s[idx-1] == ' ' {
		return true
	}
	return false
}

func validateHeaderKey(rawKey string) error {
	for _, r := range rawKey {
		if r <= 32 || r > 126 || r == ':' {
			return fmt.Errorf("invalid character in header key: %v", rawKey)
		}
	}
	return nil
}
