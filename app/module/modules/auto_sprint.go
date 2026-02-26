package modules

import (
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

var _ module.Module = (*autoSprint)(nil)

var (
	autoSprintSig   = []byte{0x0F, 0xB6, 0x41, 0x63, 0x40, 0x32, 0xED}
	autoSprintPatch = []byte{0x66, 0xB8, 0x01, 0x00, 0x40, 0x30, 0xED}
)

type AutoSprint struct {
	Process  *win.Process
	Error    func(error)
	OnChange func(bool)
}

func (conf *AutoSprint) Create(p module.Property) module.Module {
	m := &autoSprint{ToggleableModule: (&modulesutil.ByteToggleModule{
		Signature: autoSprintSig,
		Patch:     autoSprintPatch,
		Process:   conf.Process,
		Error:     conf.Error,
		OnChange:  conf.OnChange,
	}).New()}
	// sync the state
	_ = m.Set(p.Enabled)
	return m
}

// DefaultProperty ...
func (*AutoSprint) DefaultProperty() module.Property {
	return module.Property{Enabled: false}
}

// Identifier ...
func (*AutoSprint) Identifier() string {
	return "auto_sprint"
}

type autoSprint struct {
	modulesutil.ToggleableModule
}

// Name ...
func (*autoSprint) Name() string {
	return "auto sprint"
}

// Description ...
func (*autoSprint) Description() string {
	return "automatically sprints for you"
}

// Edit ...
func (a *autoSprint) Edit(p module.Property) {
	_ = a.Set(p.Enabled)
}
