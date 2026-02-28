package modules

import (
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

var _ module.Module = (*noHurtCam)(nil)

var noHurtCamSig = []byte{0x66, 0x44, 0x0F, 0x6E, 0x83, 0x6C, 0x0E, 0x00, 0x00}

type NoHurtCam struct {
	Process  *win.Process
	Error    func(error)
	OnToggle func(bool)
}

func (conf *NoHurtCam) Create(p module.Property) module.Module {
	n := &noHurtCam{ToggleableModule: (&modulesutil.SigToggleModule{
		Signature: noHurtCamSig,
		Process:   conf.Process,
		Error:     conf.Error,
		OnToggle:  conf.OnToggle,
	}).New()}
	// sync state
	_ = n.Set(p.Enabled)
	return n
}

// DefaultProperty ...
func (conf *NoHurtCam) DefaultProperty() module.Property {
	return module.Property{Enabled: false}
}

// Identifier ...
func (*NoHurtCam) Identifier() string {
	return "no_hurt_cam"
}

type noHurtCam struct {
	modulesutil.ToggleableModule
}

// Name ...
func (*noHurtCam) Name() string {
	return "no hurt cam"
}

// Description ...
func (*noHurtCam) Description() string {
	return "prevents camera shaking when player hurt"
}

// Edit ...
func (n *noHurtCam) Edit(p module.Property) {
	_ = n.Set(p.Enabled)
}
