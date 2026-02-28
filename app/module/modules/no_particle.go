package modules

import (
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

var particleSig = []byte{0xE8, 0x68, 0x4F, 0xCF, 0xFF}

type NoParticle struct {
	Process  *win.Process
	Error    func(error)
	OnToggle func(bool)
}

func (conf *NoParticle) Create(p module.Property) module.Module {
	n := &noParticle{ToggleableModule: (&modulesutil.SigToggleModule{
		Signature: particleSig,
		Process:   conf.Process,
		Error:     conf.Error,
		OnToggle:  conf.OnToggle,
	}).New()}
	// sync state
	_ = n.Set(p.Enabled)
	return n
}

// DefaultProperty ...
func (*NoParticle) DefaultProperty() module.Property {
	return module.Property{Enabled: false}
}

// Identifier ...
func (*NoParticle) Identifier() string {
	return "no_particle"
}

type noParticle struct {
	modulesutil.ToggleableModule
}

// Name ...
func (*noParticle) Name() string {
	return "no particle"
}

// Description ...
func (*noParticle) Description() string {
	return "disables particle rendering"
}

// Edit ...
func (n *noParticle) Edit(p module.Property) {
	_ = n.Set(p.Enabled) //fixme handle error
}
