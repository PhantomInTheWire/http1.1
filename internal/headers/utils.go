package headers

import "strings"

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
