package mem

type (
	OptionSubtractModule struct{}
	OptionProtect        struct{}
)

func handleOptions(opts []any) (subModule, protect bool) {
	for _, opt := range opts {
		switch opt.(type) {
		case OptionSubtractModule:
			subModule = true
		case OptionProtect:
			protect = true
		}
	}
	return
}
