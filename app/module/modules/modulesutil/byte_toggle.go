package modulesutil

import (
	"bytes"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/pkg/e"
	"github.com/something-that-is-cool/zutil/pkg/fyneutil"
	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
	"github.com/something-that-is-cool/zutil/pkg/win/mem/memutil"
)

type ByteToggleModule struct {
	Settings Settings
	Address  func(*win.Process) (uintptr, error)
	Process  *win.Process

	Error    func(error)
	OnToggle func(bool, e.ActionCause)
}

var ErrMustSetPatch = errors.New("must set patch")

func (conf ByteToggleModule) New() (t ToggleableModule, err error) {
	var addr uintptr
	if conf.Settings.Signature.Empty() {
		if conf.Address == nil {
			return nil, errors.New("empty address")
		}
		addr, err = conf.Address(conf.Process)
		if err != nil {
			return nil, fmt.Errorf("get address: %w", err)
		}
	}
	if !extendPatchFunc(&conf.Settings) {
		return nil, ErrMustSetPatch
	}
	if conf.OnToggle == nil {
		conf.OnToggle = func(bool, e.ActionCause) {}
	}
	m := &byteToggleModule{
		ErrorHandler: ErrorHandler{Error: conf.Error},
		s:            conf.Settings,
		proc:         conf.Process,
		addr:         addr,
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
			conf.OnToggle(v, cause)
			return nil
		},
	}
	m.toggler.Create()
	return m, nil
}

var _ ToggleableModule = (*byteToggleModule)(nil)

type byteToggleModule struct {
	e.ErrorHandler

	s    Settings
	proc *win.Process

	t *memutil.ByteToggler

	toggler *fyneutil.Toggler

	addr uintptr
}

func (m *byteToggleModule) UpdateState(v bool, cause e.ActionCause, opts ...any) error {
	if m.toggler.Check.Checked == v {
		return e.ErrValuesIsAlready{Value: v}
	}
	if cause == nil {
		cause = e.ActionCauseExternal
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

func (m *byteToggleModule) Edit(p module.Property, cause e.ActionCause) {
	SyncState(m, p, cause)
}

func (m *byteToggleModule) lazyToggler() (t *memutil.ByteToggler, err error) {
	if m.t != nil {
		return m.t, nil
	}
	addr := m.addr
	if !m.s.Signature.Empty() {
		addr, err = mem.ScanSignature(m.proc, m.s.Signature)
		if err != nil {
			return nil, fmt.Errorf("scan sig: %w", err)
		}
	}
	patch, err := m.extendSigWildcards(addr, m.s.Patch)
	if err != nil {
		return nil, fmt.Errorf("extend wildcards to patch: %w", err)
	}
	original, err := mem.ReadBytes(m.proc, addr, uint(len(patch)))
	if err != nil {
		return nil, fmt.Errorf("read original bytes: %w", err)
	}
	if bytes.Equal(original, patch) {
		m.t.SetState(true)
	}
	m.t = &memutil.ByteToggler{
		Process:  m.proc,
		Address:  addr,
		Original: original,
		Patch:    patch,
	}
	return m.t, nil
}

func (m *byteToggleModule) extendSigWildcards(addr uintptr, patch mem.Signature) ([]byte, error) {
	originalBytes, err := mem.ReadBytes(m.proc, addr, uint(len(patch.Data)))
	if err != nil {
		return nil, fmt.Errorf("read original bytes at 0x%X: %w", addr, err)
	}
	data := make([]byte, len(patch.Data))
	for i := 0; i < len(patch.Data); i++ {
		if i < len(patch.Mask) && patch.Mask[i] == '?' {
			data[i] = originalBytes[i]
			continue
		}
		data[i] = patch.Data[i]
	}
	return data, nil
}
