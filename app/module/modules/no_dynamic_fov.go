package modules

import (
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

var _ module.Module = (*noDynamicFov)(nil)

var noDynamicFovSig = []byte{0xF3, 0x0F, 0x11, 0x83, 0x78, 0x12, 0x00, 0x00}

type NoDynamicFov struct {
	Process  *win.Process
	Error    func(error)
	OnToggle func(bool)
}

func (conf *NoDynamicFov) Create(p module.Property) module.Module {
	n := &noDynamicFov{ToggleableModule: (&modulesutil.SigToggleModule{
		Signature: noDynamicFovSig,
		Process:   conf.Process,
		Error:     conf.Error,
		OnToggle:  conf.OnToggle,
	}).New()}
	// sync state
	_ = n.Set(p.Enabled)
	return n
}

// DefaultProperty ...
func (*NoDynamicFov) DefaultProperty() module.Property {
	return module.Property{Enabled: false}
}

// Identifier ...
func (*NoDynamicFov) Identifier() string {
	return "no_dynamic_fov"
}

type noDynamicFov struct {
	modulesutil.ToggleableModule
}

// Name ...
func (*noDynamicFov) Name() string {
	return "no dynamic fov"
}

// Description ...
func (*noDynamicFov) Description() string {
	return "forces game to think that your field of view is static"
}

// Edit ...
func (n *noDynamicFov) Edit(p module.Property) {
	_ = n.Set(p.Enabled)
}
