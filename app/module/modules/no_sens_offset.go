package modules

import (
	"fmt"
	"strings"

	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/pkg/e"
	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
)

var noSensOffsetSettings = modulesutil.Settings{
	Signature: mem.MustParseSignature("F3 0F 58 05 ?? ?? ?? ?? 80 79 19 00"),
	PatchFunc: modulesutil.PatchFuncExtendNop(8),
}

var _ module.Config = (*NoSensOffset)(nil)

type NoSensOffset struct {
	modulesutil.DefaultDisabled
	Process  *win.Process
	Error    func(error)
	OnToggle func(bool, e.ActionCause)
}

func (conf *NoSensOffset) Create(p module.Property, cause e.ActionCause) (module.Module, error) {
	cl := &sensClamper{proc: conf.Process, err: conf.Error}
	c := &modulesutil.ByteToggleModule{
		Settings: noSensOffsetSettings,
		Process:  conf.Process,
		Error:    conf.Error,
		OnToggle: func(v bool, cause e.ActionCause) {
			cl.clamp(v)
			conf.OnToggle(v, cause)
		},
	}
	b, err := c.New()
	if err != nil {
		return nil, fmt.Errorf("create byte toggle module: %w", err)
	}
	if cause == nil {
		cause = e.ActionCauseExternal
	}
	m := modulesutil.NewBaseToggleable(b,
		"no sens offset",
		"removes hardcoded 0.30000001 offset to sensitivity",
	)
	m.Edit(p, cause)
	return m, nil
}

func (*NoSensOffset) Identifier() string {
	return "no_sens_offset"
}

var sensPtr = mem.MustParsePointer(
	"019209F0",
	"10 8 0 28 A0 0 14",
)

// v4 = powf(v3 * 1.1, 1.3);
// v5 = qword_1419209F0;
// v6 = *((_QWORD **)qword_1419209F0 + 1);
// v7 = (float)(v4 * 0.41999999) + 0.30000001;
const sensOffset float32 = 0.30000001

type sensClamper struct {
	proc *win.Process
	err  func(error)

	addr uintptr
}

func (cl *sensClamper) clamp(v bool) {
	addr, ok := cl.lazyAddress()
	if !ok {
		return
	}
	old, err := mem.ReadMemory[float32](cl.proc, addr)
	if err != nil {
		cl.addr = 0
		cl.err(fmt.Errorf("read sens ptr from lazy addr: %w", err))
		return
	}
	newVal := old
	if v {
		newVal -= sensOffset
	} else {
		newVal += sensOffset
	}
	if err = mem.WriteMemory(cl.proc, addr, newVal); err != nil {
		cl.addr = 0
		cl.err(fmt.Errorf("write sens ptr %g: %w", newVal, err))
	}
}

func (cl *sensClamper) lazyAddress() (uintptr, bool) {
	if cl.addr != 0 {
		return cl.addr, true
	}
	addr, err := mem.ResolvePointerAddress(cl.proc, sensPtr)
	if err != nil {
		if !strings.Contains(err.Error(), "Only part of a ReadProcessMemory or WriteProcessMemory request was completed") {
			cl.err(fmt.Errorf("resolve sens ptr addr: %w", err))
		}
		return 0, false
	}
	cl.addr = addr
	return addr, true
}
