package modulesutil

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

type SigToggleModule struct {
	Signature []byte
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
	return func(b bool) {
		t, err := m.lazyToggler()
		if err != nil {
			m.Error(fmt.Errorf("cannot get sig toggler: %w", err))
			return
		}
		if err = t.Set(b); err != nil {
			m.Error(fmt.Errorf("cannot toggle: %w", err))
			return
		}
		if b {
			check.Text = "enabled"
			check.Checked = true
			return
		}
		check.Text = "disabled"
		check.Checked = false
	}
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
