package modulesutil

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

type SigToggleModule struct {
	Signature []byte
	Offset    uintptr
	Process   *win.Process
	Error     func(error)
	OnChange  func(bool)
}

func (conf SigToggleModule) New() ToggleableModule {
	s := &sigToggleModule{
		sig:    conf.Signature,
		offset: conf.Offset,
		proc:   conf.Process,
		err:    conf.Error,
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
		conf.OnChange(b)
		return nil
	})
	return s
}

var _ ToggleableModule = (*sigToggleModule)(nil)

type sigToggleModule struct {
	sig    []byte
	offset uintptr
	proc   *win.Process
	err    func(error)

	check *widget.Check

	t *win.SignatureNopToggler
}

// CreateObjects ...
func (m *sigToggleModule) CreateObjects() []fyne.CanvasObject {
	return []fyne.CanvasObject{m.check}
}

func (m *sigToggleModule) Set(b bool) error {
	m.check.SetChecked(b)
	return nil
}

func (m *sigToggleModule) Value() (bool, bool) {
	return m.t.Enabled(), true
}

func (m *sigToggleModule) lazyToggler() (*win.SignatureNopToggler, error) {
	if m.t != nil {
		return m.t, nil
	}
	conf := win.SignatureNopTogglerConfig{
		Process:   m.proc,
		Module:    m.proc.Module,
		Size:      m.proc.ModuleSize,
		Signature: m.sig,
	}
	t, err := conf.New()
	if err != nil {
		return nil, fmt.Errorf("init toggler: %w", err)
	}
	m.t = t
	//_ = m.toggler.Set(m.toggler.Enabled())
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
}
