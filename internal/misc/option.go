package misc

func HasTrueOption(x []bool) bool {
	for _, v := range x {
		if v {
			return true
		}
	}
	return false
}
