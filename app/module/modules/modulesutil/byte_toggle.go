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
		return e.ErrValuesIsAlready{Value: v}
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
	patch, err := m.extendSigWildcards(addr, m.sig.Patch)
	if err != nil {
		return nil, fmt.Errorf("extend wildcards to patch: %w", err)
	}
	original := m.sig.Original.Data
	if original == nil {
		original, _ = mem.ReadBytes(m.proc, addr, uint(len(patch)))
	}
	t := &memutil.ByteToggler{
		Process:  m.proc,
		Address:  addr,
		Original: original,
		Patch:    patch,
	}
	currentBytes, err := mem.ReadBytes(m.proc, addr, uint(len(patch)))
	if err == nil && bytes.Equal(currentBytes, patch) {
		t.SetState(true)
	}
	m.t = t
	return t, nil
}

func (m *byteToggleModule) extendSigWildcards(addr uintptr, patch mem.Signature) ([]byte, error) {
	originalBytes, err := mem.ReadBytes(m.proc, addr, uint(len(patch.Data)))
	if err != nil {
		return nil, fmt.Errorf("read original bytes at 0x%X: %w", addr, err)
	}
	data := make([]byte, len(patch.Data))
	copy(data, patch.Data)

	for i := 0; i < len(patch.Data); i++ {
		if i < len(m.sig.Signature.Mask) {
			if m.sig.Signature.Mask[i] == '?' {
				data[i] = originalBytes[i]
			}
		}
	}
	return data, nil
}
