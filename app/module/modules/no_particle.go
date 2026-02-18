package modules

import (
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

var particleSig = []byte{0xE8, 0x68, 0x4F, 0xCF, 0xFF}

type NoParticle struct {
	Process *win.Process
	Error   func(error)
}

func (conf NoParticle) Create() module.Module {
	return &noParticle{SigToggleModule: &modulesutil.SigToggleModule{
		Signature: particleSig,
		Process:   conf.Process,
		Error:     conf.Error,
	}}
}

type noParticle struct {
	*modulesutil.SigToggleModule
}

// Name ...
func (*noParticle) Name() string {
	return "no particle"
}

// Description ...
func (*noParticle) Description() string {
	return "disables particle rendering"
}
