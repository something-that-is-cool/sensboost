package modules

import (
	"fmt"

	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

var noDynamicFovSig = modulesutil.SignatureSettings{
	Signature: []byte{0xF3, 0x0F, 0x11, 0x83, 0x78, 0x12, 0x00, 0x00},
}

var _ module.Module = (*noDynamicFov)(nil)

type NoDynamicFov struct {
	modulesutil.DefaultDisabled
	Process  *win.Process
	Error    func(error)
	OnToggle func(bool)
}

func (conf *NoDynamicFov) Create(p module.Property) (module.Module, error) {
	c := &modulesutil.SigToggleModule{
		Sig:      noDynamicFovSig,
		Process:  conf.Process,
		Error:    conf.Error,
		OnToggle: conf.OnToggle,
	}
	t, err := c.New()
	if err != nil {
		return nil, fmt.Errorf("create sig toggle module: %w", err)
	}
	n := &noDynamicFov{ToggleableModule: t}
	_ = n.UpdateState(p.Enabled)
	return n, nil
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
	_ = n.UpdateState(p.Enabled)
}
