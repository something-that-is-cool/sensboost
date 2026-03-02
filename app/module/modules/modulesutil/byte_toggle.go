package modulesutil

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

type ByteToggleModule struct {
	Sig     SignatureSettings
	Process *win.Process

	Error    func(error)
	OnToggle func(bool)
}

func (conf ByteToggleModule) New() (ToggleableModule, error) {
	m := &byteToggleModule{
		sig:  conf.Sig,
		proc: conf.Process,
		err:  conf.Error,
	}
	m.check = &widget.Check{Text: ToggleDisabled}
	m.check.OnChanged = CheckSet(conf.Error, m.check, func(b bool, check *widget.Check) error {
		toggler, err := m.lazyToggler()
		if err != nil {
			return fmt.Errorf("get byte toggler: %w", err)
		}
		if err = toggler.Set(b); err != nil {
			return fmt.Errorf("update byte toggler state: %w", err)
		}
		conf.OnToggle(b)
		return nil
	})
	return m, nil
}

var _ ToggleableModule = (*byteToggleModule)(nil)

type byteToggleModule struct {
	sig SignatureSettings

	proc *win.Process
	err  func(error)

	t *win.ByteToggler

	check *widget.Check
}

// UpdateState ...
func (m *byteToggleModule) UpdateState(b bool) error {
	m.check.SetChecked(b)
	return nil
}

// State ...
func (m *byteToggleModule) State() bool {
	if m.t == nil {
		return false
	}
	return m.t.Enabled()
}

// CreateObjects ...
func (m *byteToggleModule) CreateObjects() []fyne.CanvasObject {
	return []fyne.CanvasObject{m.check}
}

// Disable ...
func (m *byteToggleModule) Disable() {
	if m.t == nil || !m.t.Enabled() {
		return
	}
	_ = m.t.Set(false)
}

func (m *byteToggleModule) lazyToggler() (*win.ByteToggler, error) {
	if m.t != nil {
		return m.t, nil
	}
	addr, err := win.ScanSignature(m.proc, m.proc.ModuleSize, m.proc.Module, m.sig.Signature)
	if err != nil {
		addr, err = win.ScanSignature(m.proc, m.proc.ModuleSize, m.proc.Module, m.sig.Patch)
		if err != nil {
			return nil, fmt.Errorf("signature not found: %w", err)
		}
	}
	addr += m.sig.Offset
	if m.sig.Original == nil {
		m.sig.Original = m.sig.Signature
	}
	//if len(m.Signature) < len(m.Patch) {
	// alloc for injection ?
	//}
	t := &win.ByteToggler{
		Process:  m.proc,
		Address:  addr,
		Original: m.sig.Original,
		Patch:    m.sig.Patch,
	}
	testAddr, _ := win.ScanSignature(m.proc, uintptr(len(m.sig.Patch)), addr, m.sig.Patch)
	if testAddr != 0 {
		t.SetState(true)
	}
	m.t = t
	return t, nil
}
