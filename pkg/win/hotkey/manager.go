package hotkey

import (
	"context"
	"sync"
	"sync/atomic"

	hook "github.com/robotn/gohook"
	"github.com/something-that-is-cool/zutil/internal/misc"
	"github.com/something-that-is-cool/zutil/pkg/e"
)

//fixme: gohook causes cpu spikes

type ManagerConfig struct {
	Handlers map[string]func()
}

// New ...
func (conf ManagerConfig) New() *Manager {
	if conf.Handlers == nil {
		conf.Handlers = make(map[string]func())
	}
	m := &Manager{
		Events:  misc.ValueWithRWMutex[map[string]func()]{V: make(map[string]func())},
		pressed: make(map[uint16]struct{}),
	}
	for k, h := range conf.Handlers {
		m.Events.V[k] = h
	}
	return m
}

func NewManager() *Manager {
	var conf ManagerConfig
	return conf.New()
}

type Manager struct {
	ctx    context.Context
	cancel context.CancelFunc

	closed  atomic.Bool
	running misc.ValueWithMutex[bool]

	pressed   map[uint16]struct{}
	pressedMu sync.RWMutex

	ignore atomic.Int32

	Events misc.ValueWithRWMutex[map[string]func()] //TODO: optionally call handler also on release
}

// Run ...
func (m *Manager) Run(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	m.running.Lock()
	if m.running.V {
		m.running.Unlock()
		return e.ErrAlreadyRunning
	}
	m.ctx, m.cancel = context.WithCancel(ctx)
	defer m.cancel()

	ev := hook.Start()
	defer hook.StopEvent()

	m.running.V = true
	m.running.Unlock()
	for {
		select {
		case <-m.ctx.Done():
			return m.ctx.Err()
		case ev := <-ev:
			m.handleEvent(ev)
		}
	}
}

func (m *Manager) Pressed(key string) bool {
	keycode, ok := hook.Keycode[key]
	if !ok {
		return false
	}
	m.pressedMu.RLock()
	defer m.pressedMu.RUnlock()

	_, ok = m.pressed[keycode]
	return ok
}

func (m *Manager) SkipNext(n int32) {
	m.ignore.Add(n)
}

func (m *Manager) handleEvent(ev hook.Event) {
	if m.ignore.Load() > 0 {
		m.ignore.Add(-1)
		return
	}
	switch ev.Kind {
	case hook.KeyUp:
		m.pressedMu.Lock()
		delete(m.pressed, ev.Keycode)
		m.pressedMu.Unlock()
	case hook.KeyDown:
		m.pressedMu.Lock()
		if _, ok := m.pressed[ev.Keycode]; ok {
			m.pressedMu.Unlock()
			return
		}
		m.pressed[ev.Keycode] = struct{}{}
		m.pressedMu.Unlock()
		m.triggerHandler(ev.Keycode)
	}
}

func (m *Manager) triggerHandler(keycode uint16) {
	s, ok := KeycodeToChar[keycode]
	if !ok {
		return
	}
	m.Events.RLock()
	h, ok := m.Events.V[s]
	m.Events.RUnlock()

	if !ok {
		return
	}
	h()
}

func (m *Manager) Handle(key string, fn func()) {
	if key == "" || fn == nil {
		return
	}
	m.Events.Lock()
	defer m.Events.Unlock()
	m.HandleUnsafe(key, fn)
}

func (m *Manager) HandleUnsafe(key string, fn func()) {
	m.Events.V[key] = fn
}

func (m *Manager) DeleteHandler(key string) {
	if key == "" {
		return
	}
	m.Events.Lock()
	defer m.Events.Unlock()
	m.DeleteHandlerUnsafe(key)
}

func (m *Manager) DeleteHandlerUnsafe(key string) {
	delete(m.Events.V, key)
}

func (m *Manager) ClearHandlers() {
	m.Events.Lock()
	defer m.Events.Unlock()
	m.ClearHandlersUnsafe()
}

func (m *Manager) ClearHandlersUnsafe() {
	clear(m.Events.V)
}

func (m *Manager) Close() error {
	if !m.closed.CompareAndSwap(false, true) {
		return e.ErrAlreadyClosed
	}
	m.running.Lock()
	defer m.running.Unlock()

	if !m.running.V {
		return nil
	}
	m.cancel()
	return nil
}
