package e

import (
	"errors"
	"fmt"
)

type ErrorHandler interface {
	HandleError(source string, err error)
}

var ErrAlreadyRunning = errors.New("already running")

var ErrAlreadyInitialized = errors.New("already initialized")

var ErrAlreadyClosed = errors.New("already closed")

var ErrClosed = errors.New("closed")

var ErrNotRunning = errors.New("not running")

type ErrStateAlreadyIs struct {
	State bool
}

func (err ErrStateAlreadyIs) Error() string {
	return fmt.Sprintf("state already is %t", err.State)
}

var _ ErrorHandler = (*NopErrorHandler)(nil)

type NopErrorHandler struct{}

func (NopErrorHandler) HandleError(string, error) {}
