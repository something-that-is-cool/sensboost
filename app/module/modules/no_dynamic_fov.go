package modules

import (
	"github.com/k4ties/sensboost/app/module"
	"github.com/k4ties/sensboost/app/module/modules/modulesutil"
	"github.com/k4ties/sensboost/internal/pkg/win"
)

var _ module.Module = (*noDynamicFov)(nil)

var noDynamicFovSig = []byte{0xF3, 0x0F, 0x11, 0x83, 0x78, 0x12, 0x00, 0x00}

type NoDynamicFov struct {
	Process *win.Process
	Error   func(error)
}

func (conf NoDynamicFov) Create() module.Module {
	return &noDynamicFov{SigToggleModule: &modulesutil.SigToggleModule{
		Signature: noDynamicFovSig,
		Process:   conf.Process,
		Error:     conf.Error,
	}}
}

type noDynamicFov struct {
	*modulesutil.SigToggleModule
}

// Name ...
func (*noDynamicFov) Name() string {
	return "no dynamic fov"
}

// Description ...
func (*noDynamicFov) Description() string {
	return "allows to disable dynamic field of view, that is by default always dynamic"
}
