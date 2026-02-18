package modules

import (
	"github.com/k4ties/sensboost/app/module"
	"github.com/k4ties/sensboost/app/module/modules/modulesutil"
	"github.com/k4ties/sensboost/internal/pkg/win"
)

var _ module.Module = (*noHurtCam)(nil)

var noHurtCamSig = []byte{0x66, 0x44, 0x0F, 0x6E, 0x83, 0x6C, 0x0E, 0x00, 0x00}

type NoHurtCam struct {
	Process *win.Process
	Error   func(error)
}

func (conf NoHurtCam) Create() module.Module {
	return &noHurtCam{SigToggleModule: &modulesutil.SigToggleModule{
		Signature: noHurtCamSig,
		Process:   conf.Process,
		Error:     conf.Error,
	}}
}

type noHurtCam struct {
	*modulesutil.SigToggleModule
}

// Name ...
func (*noHurtCam) Name() string {
	return "no hurt cam"
}

// Description ...
func (*noHurtCam) Description() string {
	return "allows to prevent camera shaking when player hurt"
}
