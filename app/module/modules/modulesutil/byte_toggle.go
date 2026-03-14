package modulesutil

import (
	"bytes"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
	"github.com/something-that-is-cool/zutil/pkg/win/mem/memutil"
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
		if conf.OnToggle != nil {
			conf.OnToggle(b)
		}
		return nil
	})
	return m, nil
}

var _ ToggleableModule = (*byteToggleModule)(nil)

type byteToggleModule struct {
	sig SignatureSettings

	proc *win.Process
	err  func(error)

	t *memutil.ByteToggler

	check *widget.Check
}

func (m *byteToggleModule) UpdateState(b bool) error {
	m.check.SetChecked(b)
	return nil
}

func (m *byteToggleModule) State() bool {
	if m.t == nil {
		return false
	}
	return m.t.Enabled()
}

func (m *byteToggleModule) CreateObjects() []fyne.CanvasObject {
	return []fyne.CanvasObject{m.check}
}

func (m *byteToggleModule) Disable() {
	if m.t == nil || !m.t.Enabled() {
		return
	}
	_ = m.UpdateState(false)
}

func (m *byteToggleModule) lazyToggler() (*memutil.ByteToggler, error) {
	if m.t != nil {
		return m.t, nil
	}
	addr, err := mem.ScanSignature(m.proc, m.sig.Signature)
	if err != nil {
		return nil, fmt.Errorf("signature not found: %w", err)
	}
	original := m.sig.Original
	if original == nil {
		original = m.sig.Signature.Data
	}
	t := &memutil.ByteToggler{
		Process:  m.proc,
		Address:  addr,
		Original: original,
		Patch:    m.sig.Patch.Data,
	}
	currentBytes, err := mem.ReadBytes(m.proc, addr, uint(len(m.sig.Patch.Data)))
	if err == nil && bytes.Equal(currentBytes, m.sig.Patch.Data) {
		t.SetState(true)
	}
	m.t = t
	return t, nil
}
