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

	toggler *win.SignatureNopToggler
}

// CreateObjects ...
func (m *SigToggleModule) CreateObjects() []fyne.CanvasObject {
	check := widget.NewCheck("", nil)
	check.Text = "disabled"
	check.OnChanged = m.set(check)
	return []fyne.CanvasObject{check}
}

func (m *SigToggleModule) set(check *widget.Check) func(bool) {
	return CheckSet(m.Error, check, func(b bool, check *widget.Check) error {
		toggler, err := m.lazyToggler()
		if err != nil {
			return fmt.Errorf("get sig toggler: %w", err)
		}
		if err = toggler.Set(b); err != nil {
			return fmt.Errorf("update sig toggler state: %w", err)
		}
		return nil
	})
}

func (m *SigToggleModule) lazyToggler() (*win.SignatureNopToggler, error) {
	if m.toggler != nil {
		return m.toggler, nil
	}
	conf := win.SignatureNopTogglerConfig{
		Process:   m.Process,
		Module:    m.Process.Module,
		Size:      m.Process.ModuleSize,
		Signature: m.Signature,
	}
	t, err := conf.New()
	if err != nil {
		return nil, fmt.Errorf("init toggler: %w", err)
	}
	m.toggler = t
	_ = m.toggler.Set(m.toggler.Enabled())
	return t, nil
}

func (m *SigToggleModule) Disable() {
	if m.toggler == nil || !m.toggler.Enabled() {
		return
	}
	_ = m.toggler.Set(false)
}
