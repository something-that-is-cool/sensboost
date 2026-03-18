package e

import "fmt"

type ActionCause interface {
	fmt.Stringer
	ActionCause()
}

type actionCause struct{ string }

func (a actionCause) String() string { return a.string }
func (actionCause) ActionCause()     {}

func NewActionCause(str string) ActionCause {
	return actionCause{string: str}
}

func ActionCauseIs(a, b ActionCause) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.String() == b.String()
}

var ActionCauseExternal = NewActionCause("external")
