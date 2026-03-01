package win

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"

	hook "github.com/robotn/gohook"
	"github.com/something-that-is-cool/zutil/internal/misc"
)

type HotkeyManagerConfig struct {
	Handlers map[string]func()
}

// New ...
func (conf HotkeyManagerConfig) New() *HotkeyManager {
	if conf.Handlers == nil {
		conf.Handlers = make(map[string]func())
	}
	m := &HotkeyManager{events: make(map[string]func())}
	for k, h := range conf.Handlers {
		m.events[k] = h
	}
	return m
}

type HotkeyManager struct {
	ctx    context.Context
	cancel context.CancelFunc

	l *slog.Logger

	closed  atomic.Bool
	running misc.ValueWithMutex[bool]

	events   map[string]func()
	eventsMu sync.RWMutex
}

// Run ...
func (m *HotkeyManager) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if !m.initIfStarted(ctx) {
		return ErrAlreadyRunning
	}
	defer m.cancel()
	select {
	case <-m.ctx.Done():
		return m.ctx.Err()
	case <-hook.Process(hook.Start()):
	}
	return nil
}

func (m *HotkeyManager) initIfStarted(ctx context.Context) bool {
	m.running.Lock()
	defer m.running.Unlock()

	if m.running.V {
		return false
	}
	m.ctx, m.cancel = context.WithCancel(ctx)
	m.running.V = true

	hook.Register(hook.KeyUp, nil, m.handleEvent)
	return true
}

func (m *HotkeyManager) Handle(key string, fn func()) {
	if key == "" || fn == nil {
		return
	}
	m.eventsMu.Lock()
	defer m.eventsMu.Unlock()
	m.events[key] = fn
}

func (m *HotkeyManager) DeleteHandler(key string) {
	m.eventsMu.Lock()
	defer m.eventsMu.Unlock()
	delete(m.events, key)
}

func (m *HotkeyManager) ClearHandlers() {
	m.eventsMu.Lock()
	defer m.eventsMu.Unlock()
	clear(m.events) // todo: unsafe methods + allow user to control mutex
}

var ErrAlreadyClosed = errors.New("already closed")

func (m *HotkeyManager) Close() error {
	if !m.closed.CompareAndSwap(false, true) {
		return ErrAlreadyClosed
	}
	m.running.Lock()
	defer m.running.Unlock()

	if !m.running.V {
		return nil
	}
	m.cancel()
	return nil
}

func (m *HotkeyManager) handleEvent(ev hook.Event) {
	s, ok := KeycodeToChar[ev.Keycode]
	if !ok {
		return
	}
	m.eventsMu.RLock()
	defer m.eventsMu.RUnlock()

	h, ok := m.events[s]
	if !ok {
		// event is not handled
		return
	}
	h()
}
