package app

import (
	"context"
	"errors"
)

//nolint:staticcheck // Custom errors
var (
	closeCauseExternal      = errors.New("external close cause")
	closeCauseTrackerClosed = errors.New("tracker closed")
	closeCauseContextClosed = context.Canceled
)
