package modulesutil

import (
	"fmt"
	"sync/atomic"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/mem/memutil"
)

type SigToggleModule struct {
	Sig      SignatureSettings
	Process  *win.Process
	Error    func(error)
	OnToggle func(bool)
}

func (conf SigToggleModule) New() (ToggleableModule, error) {
	s := &sigToggleModule{
		sig:  conf.Sig,
		proc: conf.Process,
		err:  conf.Error,
	}
	s.check = &widget.Check{Text: ToggleDisabled}
	s.check.OnChanged = CheckSet(conf.Error, s.check, func(v bool, _ *widget.Check) error {
		toggler, err := s.lazyToggler()
		if err != nil {
			return fmt.Errorf("get sig toggler: %w", err)
		}
		if err = toggler.Set(v); err != nil {
			return fmt.Errorf("update sig toggler state: %w", err)
		}
		if conf.OnToggle != nil {
			conf.OnToggle(v)
		}
		return nil
	})
	return s, nil
}

var _ ToggleableModule = (*sigToggleModule)(nil)

type sigToggleModule struct {
	sig SignatureSettings

	proc *win.Process
	err  func(error)

	check *widget.Check

	t atomic.Pointer[memutil.SignatureNopToggler]
}

func (m *sigToggleModule) CreateObjects() []fyne.CanvasObject {
	return []fyne.CanvasObject{m.check}
}

func (m *sigToggleModule) UpdateState(v bool) error {
	if m.check.Checked == v {
		return fmt.Errorf("state is already %t", v)
	}
	m.check.SetChecked(v)
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
	if t.Enabled() && !m.check.Checked {
		_ = m.UpdateState(true)
	}
	m.t.Store(t)
	return t, nil
}

func (m *sigToggleModule) Disable() {
	t := m.t.Load()
	if t == nil || !t.Enabled() {
		return
	}
	_ = m.UpdateState(false)
}
