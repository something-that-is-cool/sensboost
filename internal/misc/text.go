package misc

import (
	"slices"
	"strings"
)

func Rune(s string) rune {
	return []rune(s)[0]
}

func InAlphabet(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func SortByAlphabet(x []string) {
	slices.SortFunc(x, func(a, b string) int {
		if pa, pb := CompareStringAlphabet(a), CompareStringAlphabet(b); pa != pb {
			return pa - pb
		}
		return strings.Compare(a, b)
	})
}

func CompareStringAlphabet(s string) int {
	if len(s) != 1 {
		return len(s) + 2
	}
	if InAlphabet(Rune(s)) {
		return len(s) + 1
	}
	return len(s) + 3
}
