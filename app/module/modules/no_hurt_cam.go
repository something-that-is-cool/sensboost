package modules

import (
	"fmt"

	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
)

var noHurtCamSig = modulesutil.SignatureSettings{
	Signature: mem.MustParseSignature("66 44 0F 6E 83 6C 0E 00 00"),
}

var _ module.Config = (*NoHurtCam)(nil)

type NoHurtCam struct {
	modulesutil.DefaultDisabled
	Process  *win.Process
	Error    func(error)
	OnToggle func(bool)
}

func (conf *NoHurtCam) Create(p module.Property) (module.Module, error) {
	c := &modulesutil.SigToggleModule{
		Sig:      noHurtCamSig,
		Process:  conf.Process,
		Error:    conf.Error,
		OnToggle: conf.OnToggle,
	}
	s, err := c.New()
	if err != nil {
		return nil, fmt.Errorf("create nop sig toggle module: %w", err)
	}
	n := &noHurtCam{ToggleableModule: s}
	n.Edit(p)
	return n, nil
}

// Identifier ...
func (*NoHurtCam) Identifier() string {
	return "no_hurt_cam"
}

var _ module.Module = (*noHurtCam)(nil)

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
	_ = n.UpdateState(p.Enabled)
}
