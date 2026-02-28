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
	OnToggle  func(bool)
}

func (conf ByteToggleModule) New() ToggleableModule {
	m := &byteToggleModule{
		sig:    conf.Signature,
		orig:   conf.Original,
		patch:  conf.Patch,
		offset: conf.Offset,
		proc:   conf.Process,
		err:    conf.Error,
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
	return m
}

var _ ToggleableModule = (*byteToggleModule)(nil)

type byteToggleModule struct {
	sig, orig, patch []byte
	offset           uintptr

	proc *win.Process
	err  func(error)

	t *win.ByteToggler

	check *widget.Check
}

func (m *byteToggleModule) Set(b bool) error {
	m.check.SetChecked(b)
	return nil
}

func (m *byteToggleModule) Value() (bool, bool) {
	if m.t == nil {
		return false, false
	}
	return m.t.Enabled(), true
}

func (m *byteToggleModule) CreateObjects() []fyne.CanvasObject {
	return []fyne.CanvasObject{m.check}
}

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
	addr, err := win.ScanSignature(m.proc, m.proc.ModuleSize, m.proc.Module, m.sig)
	if err != nil {
		addr, err = win.ScanSignature(m.proc, m.proc.ModuleSize, m.proc.Module, m.patch)
		if err != nil {
			return nil, fmt.Errorf("signature not found: %w", err)
		}
	}
	addr += m.offset
	if len(m.orig) == 0 {
		m.orig = m.sig
	}
	//if len(m.Signature) < len(m.Patch) {
	// alloc for injection ?
	//}
	t := &win.ByteToggler{
		Process:  m.proc,
		Address:  addr,
		Original: m.orig,
		Patch:    m.patch,
	}
	testAddr, _ := win.ScanSignature(m.proc, uintptr(len(m.patch)), addr, m.patch)
	if testAddr != 0 {
		t.SetState(true)
	}
	m.t = t
	return t, nil
}
