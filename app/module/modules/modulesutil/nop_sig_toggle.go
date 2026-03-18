package modulesutil

import (
	"fmt"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"github.com/something-that-is-cool/zutil/pkg/e"
	"github.com/something-that-is-cool/zutil/pkg/fyneutil"
	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/mem/memutil"
)

type SigToggleModule struct {
	Sig      SignatureSettings
	Process  *win.Process
	Error    func(error)
	OnToggle func(bool, e.ActionCause)
}

func (conf SigToggleModule) New() (ToggleableModule, error) {
	s := &sigToggleModule{
		ErrorHandler: errorHandler{err: conf.Error},
		sig:          conf.Sig,
		proc:         conf.Process,
	}
	s.toggler = &fyneutil.Toggler{
		Handler: s,
		Action: func(v bool, cause e.ActionCause) error {
			toggler, err := s.lazyToggler()
			if err != nil {
				return fmt.Errorf("get sig toggler: %w", err)
			}
			if err = toggler.Set(v); err != nil {
				return fmt.Errorf("update sig toggler state: %w", err)
			}
			if conf.OnToggle != nil {
				conf.OnToggle(v, cause)
			}
			return nil
		},
	}
	s.toggler.Create()
	return s, nil
}

var _ ToggleableModule = (*sigToggleModule)(nil)

type sigToggleModule struct {
	e.ErrorHandler

	sig  SignatureSettings
	proc *win.Process

	toggler *fyneutil.Toggler

	t atomic.Pointer[memutil.SignatureNopToggler]
}

func (m *sigToggleModule) CreateObjects() []fyne.CanvasObject {
	return []fyne.CanvasObject{m.toggler.Check}
}

func (m *sigToggleModule) UpdateState(v bool, cause e.ActionCause, opts ...any) error {
	if m.State() == v {
		return e.ErrValuesIsAlready{Value: v}
	}
	if cause == nil {
		cause = e.ActionCauseExternal
	}
	m.toggler.Set(v, cause, opts...)
	return nil
}

func (m *sigToggleModule) State() bool {
	t := m.t.Load()
	if t == nil {
		return false
	}
	return t.Enabled()
}

func (m *sigToggleModule) lazyToggler() (*memutil.SignatureNopToggler, error) {
	if t := m.t.Load(); t != nil {
		return t, nil
	}
	conf := memutil.SignatureNopTogglerConfig{
		Process:   m.proc,
		Signature: m.sig.Signature,
	}
	t, err := conf.New()
	if err != nil {
		return nil, fmt.Errorf("init toggler: %w", err)
	}
	if t.Enabled() && !m.toggler.Check.Checked {
		m.toggler.Check.SetChecked(true)
		m.toggler.Check.Text = fyneutil.ToggleEnabled
	}
	m.t.Store(t)
	return t, nil
}

func (m *sigToggleModule) Disable(cause e.ActionCause) {
	m.HandleError("disable sig toggle module", disableOnlyAction(m, cause))
}
