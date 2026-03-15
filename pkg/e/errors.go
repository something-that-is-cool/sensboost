package e

import "errors"

var ErrAlreadyRunning = errors.New("already running")

var ErrAlreadyInitialized = errors.New("already initialized")

var ErrAlreadyClosed = errors.New("already closed")

var ErrClosed = errors.New("closed")

var ErrNotRunning = errors.New("not running")
