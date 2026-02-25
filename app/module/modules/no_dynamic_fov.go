package modules

import (
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
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
	return "forces game to think that your field of view is static"
}

// Identifier ...
func (*noDynamicFov) Identifier() string {
	return "no_dynamic_fov"
}
