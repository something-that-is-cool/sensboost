package e

import "errors"

type CloseCause interface {
	error
	CloseCause()
}

type closeCause struct{ e error }

func (c closeCause) Error() string { return c.e.Error() }
func (closeCause) CloseCause()     {}

func NewCloseCause(err error) CloseCause {
	return closeCause{e: err}
}

func NewCloseCauseString(s string) CloseCause {
	return NewCloseCause(errors.New(s))
}

func CloseCauseIs(a, b CloseCause) bool {
	return errors.Is(a, b)
}

var CloseCauseExternal = NewCloseCauseString("external")
