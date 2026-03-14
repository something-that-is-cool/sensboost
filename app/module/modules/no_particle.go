package modules

import (
	"fmt"

	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
)

var noParticleSig = modulesutil.SignatureSettings{
	Signature: mem.MustParseSignature("E8 68 4F CF FF"),
}

var _ module.Config = (*NoParticle)(nil)

type NoParticle struct {
	modulesutil.DefaultDisabled
	Process  *win.Process
	Error    func(error)
	OnToggle func(bool)
}

func (conf *NoParticle) Create(p module.Property) (module.Module, error) {
	c := &modulesutil.SigToggleModule{
		Sig:      noParticleSig,
		Process:  conf.Process,
		Error:    conf.Error,
		OnToggle: conf.OnToggle,
	}
	s, err := c.New()
	if err != nil {
		return nil, fmt.Errorf("create sig toggle module: %w", err)
	}
	n := &noParticle{ToggleableModule: s}
	n.Edit(p)
	return n, nil
}

// Identifier ...
func (*NoParticle) Identifier() string {
	return "no_particle"
}

var _ module.Module = (*noParticle)(nil)

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
	_ = n.UpdateState(p.Enabled) //todo: handle error
}
