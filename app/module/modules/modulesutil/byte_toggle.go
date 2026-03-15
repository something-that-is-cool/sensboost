package modulesutil

import (
	"bytes"
	"fmt"

	"fyne.io/fyne/v2"
	"github.com/something-that-is-cool/zutil/pkg/e"
	"github.com/something-that-is-cool/zutil/pkg/fyneutil"
	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
	"github.com/something-that-is-cool/zutil/pkg/win/mem/memutil"
)

type ByteToggleModule struct {
	Sig     SignatureSettings
	Process *win.Process

	Error    func(error)
	OnToggle func(bool, e.ActionCause)
}

func (conf ByteToggleModule) New() (ToggleableModule, error) {
	m := &byteToggleModule{
		ErrorHandler: errorHandler{err: conf.Error},
		sig:          conf.Sig,
		proc:         conf.Process,
	}
	m.toggler = &fyneutil.Toggler{
		Handler: m,
		Action: func(v bool, cause e.ActionCause) error {
			toggler, err := m.lazyToggler()
			if err != nil {
				return fmt.Errorf("get byte toggler: %w", err)
			}
			if err = toggler.Set(v); err != nil {
				return fmt.Errorf("update byte toggler state: %w", err)
			}
			if conf.OnToggle != nil {
				conf.OnToggle(v, cause)
			}
			return nil
		},
	}
	m.toggler.Create()
	return m, nil
}

var _ ToggleableModule = (*byteToggleModule)(nil)

type byteToggleModule struct {
	e.ErrorHandler

	sig  SignatureSettings
	proc *win.Process

	t *memutil.ByteToggler

	toggler *fyneutil.Toggler
}

func (m *byteToggleModule) UpdateState(v bool, cause e.ActionCause, opts ...any) error {
	if m.toggler.Check.Checked == v {
		return &e.ErrValuesIsAlready{Value: v}
	}
	m.toggler.Set(v, cause, opts...)
	return nil
}

func (m *byteToggleModule) State() bool {
	if m.t == nil {
		return false
	}
	return m.t.Enabled()
}

func (m *byteToggleModule) CreateObjects() []fyne.CanvasObject {
	return []fyne.CanvasObject{m.toggler.Check}
}

func (m *byteToggleModule) Disable(cause e.ActionCause) {
	m.HandleError("disable byte toggle module", disableOnlyAction(m, cause))
}

func (m *byteToggleModule) lazyToggler() (*memutil.ByteToggler, error) {
	if m.t != nil {
		return m.t, nil
	}
	addr, err := mem.ScanSignature(m.proc, m.sig.Signature)
	if err != nil {
		return nil, fmt.Errorf("scan sig: %w", err)
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
