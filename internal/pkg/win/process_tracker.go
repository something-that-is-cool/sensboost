package win

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	w "golang.org/x/sys/windows"
)

type ProcessTrackerConfig struct {
	Handlers []func()
	Process  *Process
}

func (conf ProcessTrackerConfig) New() (*ProcessTracker, error) {
	if conf.Process == nil {
		return nil, errors.New("nil process")
	}
	return &ProcessTracker{pr: conf.Process, handlers: conf.Handlers}, nil
}

type ProcessTracker struct {
	pr *Process

	handlers []func()

	ctx    context.Context
	cancel context.CancelFunc

	closed, running atomic.Bool
}

func (tr *ProcessTracker) Process() *Process {
	return tr.pr
}

func (tr *ProcessTracker) Close() bool {
	if !tr.closed.CompareAndSwap(false, true) {
		return false
	}
	tr.cancel()
	return true
}

var ErrTrackerClosed = errors.New("tracker closed")

var ErrAlreadyRunning = errors.New("already running")

func (tr *ProcessTracker) Run(parent context.Context) (err error) {
	if tr.closed.Load() {
		return ErrTrackerClosed
	}
	if !tr.running.CompareAndSwap(false, true) {
		return ErrAlreadyRunning
	}
	select {
	case <-parent.Done():
		return parent.Err()
	default:
	}
	if tr.pr.Handle == w.InvalidHandle {
		tr.Close()
		return errors.New("invalid process handle")
	}
	tr.ctx, tr.cancel = context.WithCancel(parent)
	defer tr.cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	err = tr.loop(ticker)
	for _, fn := range tr.handlers {
		fn()
	}
	return
}

func (tr *ProcessTracker) loop(ticker *time.Ticker) error {
	for {
		select {
		case <-tr.ctx.Done():
			return tr.ctx.Err()
		case <-ticker.C:
			if !tr.pr.Active() {
				return nil
			}
		}
	}
}

func (tr *ProcessTracker) CloseWithProcess() error {
	if !tr.Close() {
		return ErrTrackerClosed
	}
	if err := tr.pr.Close(); err != nil {
		return fmt.Errorf("close process: %w", err)
	}
	return nil
}
