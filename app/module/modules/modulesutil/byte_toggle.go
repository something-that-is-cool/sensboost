package modulesutil

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

type ByteToggleModule struct {
	Signature []byte
	Patch     []byte
	Process   *win.Process
	Error     func(error)

	toggler *win.ByteToggler
}

// CreateObjects ...
func (m *ByteToggleModule) CreateObjects() []fyne.CanvasObject {
	check := widget.NewCheck("", nil)
	check.Text = "disabled"
	check.OnChanged = m.set(check)
	return []fyne.CanvasObject{check}
}

func (m *ByteToggleModule) set(check *widget.Check) func(bool) {
	return func(b bool) {
		t, err := m.lazyToggler()
		if err != nil {
			m.Error(fmt.Errorf("cannot get byte toggler: %w", err))
			return
		}
		if err = t.Set(b); err != nil {
			m.Error(fmt.Errorf("cannot toggle: %w", err))
			return
		}
		if b {
			check.Text = "enabled"
			check.Checked = true
		} else {
			check.Text = "disabled"
			check.Checked = false
		}
		check.Refresh()
	}
}

func (m *ByteToggleModule) lazyToggler() (*win.ByteToggler, error) {
	if m.toggler != nil {
		return m.toggler, nil
	}
	addr, err := win.ScanSignature(m.Process, m.Process.ModuleSize, m.Process.Module, m.Signature)
	if err != nil {
		addr, err = win.ScanSignature(m.Process, m.Process.ModuleSize, m.Process.Module, m.Patch)
		if err != nil {
			return nil, fmt.Errorf("signature not found: %w", err)
		}
	}
	t := &win.ByteToggler{
		Process:  m.Process,
		Address:  addr,
		Original: m.Signature,
		Patch:    m.Patch,
	}
	testAddr, _ := win.ScanSignature(m.Process, uintptr(len(m.Patch)), addr, m.Patch)
	if testAddr != 0 {
		t.SetState(true)
	}
	m.toggler = t
	return t, nil
}

func (m *ByteToggleModule) Disable() {
	if m.toggler == nil || !m.toggler.Enabled() {
		return
	}
	_ = m.toggler.Set(false)
}
