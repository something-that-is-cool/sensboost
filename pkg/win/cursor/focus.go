package cursor

func Focused() bool {
	i, err := getInfo()
	if err != nil {
		return false
	}
	return i.Flags == 0
}
