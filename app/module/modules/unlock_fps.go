package modules

import (
	"fmt"

	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

var unlockFPSSignature = modulesutil.SignatureSettings{
	Signature: []byte{0x41, 0x8B, 0x9D, 0x98, 0x00, 0x00, 0x00},
}

var _ module.Module = (*unlockFPS)(nil)

type UnlockFPS struct {
	modulesutil.DefaultDisabled
	Process  *win.Process
	Error    func(error)
	OnToggle func(bool)
}

// Create ...
func (conf *UnlockFPS) Create(p module.Property) (module.Module, error) {
	c := &modulesutil.SigToggleModule{
		Sig:      unlockFPSSignature,
		Process:  conf.Process,
		Error:    conf.Error,
		OnToggle: conf.OnToggle,
	}
	b, err := c.New()
	if err != nil {
		return nil, fmt.Errorf("create byte toggler module: %w", err)
	}
	m := &unlockFPS{ToggleableModule: b}
	_ = m.UpdateState(p.Enabled)
	return m, nil
}

// Identifier ...
func (conf *UnlockFPS) Identifier() string {
	return "unlock_fps"
}

type unlockFPS struct {
	modulesutil.ToggleableModule
}

// Name ...
func (u *unlockFPS) Name() string {
	return "unlock fps"
}

// Description ...
func (u *unlockFPS) Description() string {
	return "..."
}

// Edit ...
func (u *unlockFPS) Edit(property module.Property) {
	_ = u.UpdateState(property.Enabled)
}
