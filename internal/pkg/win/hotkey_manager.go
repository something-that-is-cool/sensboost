package win

import (
	"context"
	"errors"
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
	m := &HotkeyManager{Events: misc.ValueWithRWMutex[map[string]func()]{V: make(map[string]func())}}
	for k, h := range conf.Handlers {
		m.Events.V[k] = h
	}
	return m
}

type HotkeyManager struct {
	ctx    context.Context
	cancel context.CancelFunc

	closed  atomic.Bool
	running misc.ValueWithMutex[bool]

	Events misc.ValueWithRWMutex[map[string]func()]
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
	m.Events.Lock()
	defer m.Events.Unlock()
	m.HandleUnsafe(key, fn)
}

func (m *HotkeyManager) HandleUnsafe(key string, fn func()) {
	m.Events.V[key] = fn
}

func (m *HotkeyManager) DeleteHandler(key string) {
	if key == "" {
		return
	}
	m.Events.Lock()
	defer m.Events.Unlock()
	m.DeleteHandlerUnsafe(key)
}

func (m *HotkeyManager) DeleteHandlerUnsafe(key string) {
	delete(m.Events.V, key)
}

func (m *HotkeyManager) ClearHandlers() {
	m.Events.Lock()
	defer m.Events.Unlock()
	m.ClearHandlersUnsafe()
}

func (m *HotkeyManager) ClearHandlersUnsafe() {
	clear(m.Events.V)
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
	m.Events.RLock()
	defer m.Events.RUnlock()

	h, ok := m.Events.V[s]
	if !ok {
		// event is not handled
		return
	}
	h()
}
