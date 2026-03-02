package modules

import (
	"fmt"

	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

var noFireSig = modulesutil.SignatureSettings{
	Signature: []byte{0x48, 0x89, 0x42, 0x10, 0x48, 0x83, 0xC1, 0x18},
	Patch:     []byte{0x90, 0x90, 0x90, 0x90, 0x48, 0x83, 0xC1, 0x18}, // keep add rcx,18
}

// mov [rdx+10],rax <--
// add rcx,18

var _ module.Module = (*noFire)(nil)

type NoFire struct {
	modulesutil.DefaultDisabled
	Process  *win.Process
	Error    func(error)
	OnToggle func(bool)
}

// Create ...
func (conf *NoFire) Create(p module.Property) (module.Module, error) {
	c := &modulesutil.ByteToggleModule{
		Sig:     noFireSig,
		Process: conf.Process,
		Error:   conf.Error,
		OnToggle: func(b bool) {
			onNoFireToggle(b, conf.Process, conf.Error)
			conf.OnToggle(b)
		},
	}
	b, err := c.New()
	if err != nil {
		return nil, fmt.Errorf("create byte toggler module: %w", err)
	}
	m := &noFire{ToggleableModule: b}
	_ = m.UpdateState(p.Enabled)
	return m, nil
}

// Identifier ...
func (conf *NoFire) Identifier() string {
	return "no_fire"
}

type noFire struct {
	modulesutil.ToggleableModule
}

// Name ...
func (n *noFire) Name() string {
	return "no fire"
}

// Description ...
func (n *noFire) Description() string {
	return "prevents game updating onFire flag, force writing 0 (false) when toggling module"
}

// Edit ...
func (n *noFire) Edit(property module.Property) {
	_ = n.UpdateState(property.Enabled)
}

var (
	onFirePtrBase uintptr = 0x01921DF8
	onFireOffsets         = []uintptr{0x18, 0x60, 0xF0, 0x0, 0x10}
)

func onNoFireToggle(b bool, proc *win.Process, error func(error)) {
	if !b {
		return
	}
	mod, _, err := proc.GetModuleInfo()
	if err != nil {
		error(fmt.Errorf("get process module info: %w", err))
		return
	}
	// force write false to onFire flag
	addr, err := win.ResolvePointerAddress(proc, mod, onFirePtrBase, onFireOffsets)
	if err != nil {
		error(fmt.Errorf("resolve onFire flag pointer address: %w", err))
		return
	}
	err = win.WriteMemory[byte](proc, addr, 0)
	if err != nil {
		error(fmt.Errorf("write onFire flag 0: %w", err))
	}
}
