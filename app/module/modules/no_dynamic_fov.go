package modules

import (
	"fmt"

	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/pkg/e"
	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
)

var noDynamicFovSig = modulesutil.SignatureSettings{
	Signature: mem.MustParseSignature("F3 0F 11 83 78 12 00 00"),
}
var fovPtr = modulesutil.PointerSettings{
	BaseAddress: 0x01921DF8, //+module
	Offsets:     []uintptr{0x30, 0xD8, 0x20, 0xBE0},
}

var _ module.Config = (*NoDynamicFov)(nil)

type NoDynamicFov struct {
	modulesutil.DefaultDisabled
	Process  *win.Process
	Error    func(error)
	OnToggle func(bool, e.ActionCause)
}

func (conf *NoDynamicFov) Create(p module.Property, cause e.ActionCause) (module.Module, error) {
	c := &modulesutil.SigToggleModule{
		Sig:     noDynamicFovSig,
		Process: conf.Process,
		Error:   conf.Error,
	}
	cl := &fovClamper{proc: conf.Process, err: conf.Error}
	c.OnToggle = func(b bool, cause e.ActionCause) {
		cl.clamp()
		conf.OnToggle(b, cause)
	}
	t, err := c.New()
	if err != nil {
		return nil, fmt.Errorf("create sig toggle module: %w", err)
	}
	if cause == nil {
		cause = e.ActionCauseExternal
	}
	n := &noDynamicFov{ToggleableModule: t}
	n.Edit(p, cause)
	return n, nil
}

// Identifier ...
func (*NoDynamicFov) Identifier() string {
	return "no_dynamic_fov"
}

var _ module.Module = (*noDynamicFov)(nil)

type noDynamicFov struct {
	modulesutil.ToggleableModule
}

// Name ...
func (*noDynamicFov) Name() string {
	return "no dynamic fov"
}

// Description ...
func (*noDynamicFov) Description() string {
	return "forces game to think that your field of view is static"
}

// Edit ...
func (n *noDynamicFov) Edit(p module.Property, cause e.ActionCause) {
	modulesutil.SyncState(n, p, cause)
}

const maxStaticFloatValue float32 = 1.0

type fovClamper struct {
	proc *win.Process
	err  func(error)

	addr uintptr
}

func (m *fovClamper) clamp() {
	addr, ok := m.lazyAddress()
	if !ok {
		return
	}
	v, err := mem.ReadMemory[float32](m.proc, addr)
	if err != nil {
		m.addr = 0 // force recalculate address
		m.err(fmt.Errorf("read fov ptr from lazy address: %w", err))
		return
	}
	if v <= maxStaticFloatValue {
		return
	}
	// force write max static fov value
	if err = mem.WriteMemory(m.proc, addr, maxStaticFloatValue); err != nil {
		m.err(fmt.Errorf("write fov ptr (current value overflows max static value): %w", err))
	}
}

func (m *fovClamper) lazyAddress() (uintptr, bool) {
	if m.addr != 0 {
		return m.addr, true
	}
	addr, err := mem.ResolvePointerAddress(m.proc, fovPtr.BaseAddress, fovPtr.Offsets)
	if err != nil {
		m.err(fmt.Errorf("resolve fov ptr addr: %w", err))
		return 0, false
	}
	m.addr = addr
	return addr, true
}
