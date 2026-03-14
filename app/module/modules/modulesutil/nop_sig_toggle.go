package modulesutil

import (
	"fmt"

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
	s.check.OnChanged = CheckSet(conf.Error, s.check, func(b bool, check *widget.Check) error {
		toggler, err := s.lazyToggler()
		if err != nil {
			return fmt.Errorf("get sig toggler: %w", err)
		}
		if err = toggler.Set(b); err != nil {
			return fmt.Errorf("update sig toggler state: %w", err)
		}
		if conf.OnToggle != nil {
			conf.OnToggle(b)
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

	t *memutil.SignatureNopToggler
}

func (m *sigToggleModule) CreateObjects() []fyne.CanvasObject {
	return []fyne.CanvasObject{m.check}
}

func (m *sigToggleModule) UpdateState(b bool) error {
	m.check.SetChecked(b)
	return nil
}

func (m *sigToggleModule) State() bool {
	if m.t == nil {
		return false
	}
	return m.t.Enabled()
}

func (m *sigToggleModule) lazyToggler() (*memutil.SignatureNopToggler, error) {
	if m.t != nil {
		return m.t, nil
	}
	conf := memutil.SignatureNopTogglerConfig{
		Process:   m.proc,
		Signature: m.sig.Signature,
	}
	t, err := conf.New()
	if err != nil {
		return nil, fmt.Errorf("init toggler: %w", err)
	}
	m.t = t
	if t.Enabled() {
		_ = m.UpdateState(true)
	}
	return t, nil
}

func (m *sigToggleModule) Disable() {
	if m.t == nil || !m.t.Enabled() {
		return
	}
	err := m.t.Set(false)
	if err != nil {
		m.err(fmt.Errorf("disable (set false): %w", err))
	}
	_ = m.UpdateState(false)
}
