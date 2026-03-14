package win

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/something-that-is-cool/zutil/internal/misc"
	"github.com/something-that-is-cool/zutil/pkg/e"
	w "golang.org/x/sys/windows"
)

type ProcessTrackerConfig struct {
	CloseHandlers []func()
	Process       *Process
}

func (conf ProcessTrackerConfig) New() (*ProcessTracker, error) {
	if conf.Process == nil {
		return nil, errors.New("nil process")
	}
	return &ProcessTracker{pr: conf.Process, handlers: conf.CloseHandlers}, nil
}

type ProcessTracker struct {
	pr *Process

	handlers []func()

	ctx    context.Context
	cancel context.CancelFunc

	closed  atomic.Bool
	running misc.ValueWithMutex[bool]
}

func (tr *ProcessTracker) Process() *Process {
	return tr.pr
}

func (tr *ProcessTracker) Close() bool {
	if !tr.closed.CompareAndSwap(false, true) {
		return false //e.ErrAlreadyClosed
	}
	tr.running.Lock()
	defer tr.running.Unlock()

	if !tr.running.V {
		return true
	}
	tr.cancel()
	return true
}

func (tr *ProcessTracker) Run(parent context.Context) error {
	if tr.closed.Load() {
		return e.ErrClosed
	}
	if tr.pr.Handle == w.InvalidHandle {
		tr.Close()
		return errors.New("invalid process handle")
	}
	select {
	case <-parent.Done():
		return parent.Err()
	default:
	}
	tr.running.Lock()
	if tr.running.V {
		tr.running.Unlock()
		return e.ErrAlreadyRunning
	}
	tr.ctx, tr.cancel = context.WithCancel(parent)
	defer tr.cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	tr.running.V = true
	tr.running.Unlock()

	err := tr.loop(ticker)
	// call handlers when loop returned
	for _, fn := range tr.handlers {
		fn()
	}
	return err
}

func (tr *ProcessTracker) loop(ticker *time.Ticker) error {
	for {
		select {
		case <-tr.ctx.Done():
			return tr.ctx.Err()
		case <-ticker.C: //fixme find a cheap way to instantly detect if process killed
			if !tr.pr.Active() {
				return nil
			}
		}
	}
}

func (tr *ProcessTracker) CloseWithProcess() error {
	if !tr.Close() {
		return e.ErrClosed
	}
	if err := tr.pr.Close(); err != nil {
		return fmt.Errorf("close process: %w", err)
	}
	return nil
}
