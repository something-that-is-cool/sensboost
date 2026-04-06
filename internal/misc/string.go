package misc

import "strings"

func JoinNewLine(str ...string) string {
	return strings.Join(str, "\n")
}

func CompareString(a, b string, caseSensitive ...bool) bool {
	if !HasTrueOption(caseSensitive) {
		return strings.EqualFold(a, b)
	}
	return a == b
}
