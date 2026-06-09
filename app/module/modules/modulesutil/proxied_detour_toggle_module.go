package modulesutil

import (
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/pkg/e"
	"github.com/something-that-is-cool/zutil/pkg/fyneutil"
	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
	"github.com/something-that-is-cool/zutil/pkg/win/mem/memutil"
)

type ProxiedDetourToggleModule[T any] struct {
	Settings   Settings
	Address    uintptr
	TargetSize uint
	Process    *win.Process

	UserCode     func(valAddr uintptr) []byte
	DefaultValue T

	Error    func(error)
	OnToggle func(bool, e.ActionCause)
}

func (conf ProxiedDetourToggleModule[T]) New() (t ToggleableModule, err error) {
	if conf.UserCode == nil {
		return nil, errors.New("empty user code")
	}
	if conf.Address <= 0 && conf.Settings.Signature.Empty() {
		return nil, errors.New("empty address")
	}
	if conf.OnToggle == nil {
		conf.OnToggle = func(bool, e.ActionCause) {}
	}
	addr := conf.Address
	if addr <= 0 {
		addr, err = mem.ScanSignature(conf.Process, conf.Settings.Signature)
		if err != nil {
			return nil, fmt.Errorf("scan sig: %w", err)
		}
	}
	m := &proxiedDetourToggleModule[T]{
		ErrorHandler: ErrorHandler{Error: conf.Error},
		proc:         conf.Process,
		det:          memutil.NewProxiedDetour[T](conf.Process, addr+conf.Settings.Offset, conf.TargetSize),
		uc:           conf.UserCode,
		def:          conf.DefaultValue,
	}
	act := func(v bool, cause e.ActionCause) error {
		if !v {
			if err := m.det.Disable(); err != nil {
				return fmt.Errorf("disable proxied detour: %w", err)
			}
			conf.OnToggle(v, cause)
			return nil
		}
		if err := m.det.Enable(m.uc, m.def); err != nil {
			return fmt.Errorf("enable proxied detour: %w", err)
		}
		conf.OnToggle(v, cause)
		return nil
	}
	m.toggler = &fyneutil.Toggler{
		Handler: m,
		Action:  act,
	}
	m.toggler.Create()
	return m, nil
}

type proxiedDetourToggleModule[T any] struct {
	e.ErrorHandler
	proc    *win.Process
	det     *memutil.ProxiedDetour[T]
	toggler *fyneutil.Toggler
	uc      func(uintptr) []byte
	def     T
}

func (m *proxiedDetourToggleModule[T]) UpdateState(v bool, cause e.ActionCause, opts ...any) error {
	if m.toggler.Check.Checked == v {
		return e.ErrValuesIsAlready{Value: v}
	}
	m.toggler.Set(v, cause, opts...)
	return nil
}

func (m *proxiedDetourToggleModule[T]) State() bool {
	return m.toggler.Check.Checked
}

func (m *proxiedDetourToggleModule[T]) CreateObjects() []fyne.CanvasObject {
	return []fyne.CanvasObject{m.toggler.Check}
}

func (m *proxiedDetourToggleModule[T]) Disable(cause e.ActionCause) {
	m.HandleError("disable proxied detour module", disableOnlyAction(m, cause))
}

func (m *proxiedDetourToggleModule[T]) Edit(p module.Property, cause e.ActionCause) {
	SyncState(m, p, cause)
}
