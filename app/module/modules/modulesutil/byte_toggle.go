package modulesutil

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

type ByteToggleModule struct {
	Signature []byte
	Offset    uintptr //optional
	Original  []byte  //optional
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
	return CheckSet(m.Error, check, func(b bool, check *widget.Check) error {
		toggler, err := m.lazyToggler()
		if err != nil {
			return fmt.Errorf("get byte toggler: %w", err)
		}
		if err = toggler.Set(b); err != nil {
			return fmt.Errorf("update byte toggler state: %w", err)
		}
		return nil
	})
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
	addr += m.Offset
	if len(m.Original) == 0 {
		m.Original = m.Signature
	}
	//if len(m.Signature) < len(m.Patch) {
	// alloc for injection ?
	//}
	t := &win.ByteToggler{
		Process:  m.Process,
		Address:  addr,
		Original: m.Original,
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
