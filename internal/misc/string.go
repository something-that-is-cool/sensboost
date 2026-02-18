package misc

import "strings"

func JoinNewLine(str ...string) string {
	return strings.Join(str, "\n")
}
