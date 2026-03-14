package modules

import (
	"fmt"

	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
)

var noDynamicFovSig = modulesutil.SignatureSettings{
	Signature: mem.MustParseSignature("F3 0F 11 83 78 12 00 00"),
}

var _ module.Config = (*NoDynamicFov)(nil)

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
	n.Edit(p)
	return n, nil
}

// Identifier ...
func (*NoDynamicFov) Identifier() string {
	return "no_dynamic_fov"
}

var _ module.Module = (*noDynamicFov)(nil)

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
