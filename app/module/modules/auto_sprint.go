package modules

import (
	"fmt"

	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

var autoSprintSig = modulesutil.SignatureSettings{
	Signature: []byte{0x0F, 0xB6, 0x41, 0x63, 0x40, 0x32, 0xED},
	Patch:     []byte{0x66, 0xB8, 0x01, 0x00, 0x40, 0x30, 0xED},
}

var _ module.Module = (*autoSprint)(nil)

type AutoSprint struct {
	modulesutil.DefaultDisabled
	Process  *win.Process
	Error    func(error)
	OnToggle func(bool)
}

func (conf *AutoSprint) Create(p module.Property) (module.Module, error) {
	c := &modulesutil.ByteToggleModule{
		Sig:      autoSprintSig,
		Process:  conf.Process,
		Error:    conf.Error,
		OnToggle: conf.OnToggle,
	}
	b, err := c.New()
	if err != nil {
		return nil, fmt.Errorf("create byte toggle module: %w", err)
	}
	m := &autoSprint{ToggleableModule: b}
	_ = m.UpdateState(p.Enabled)
	return m, nil
}

// Identifier ...
func (*AutoSprint) Identifier() string {
	return "auto_sprint"
}

type autoSprint struct {
	modulesutil.ToggleableModule
}

// Name ...
func (*autoSprint) Name() string {
	return "auto sprint"
}

// Description ...
func (*autoSprint) Description() string {
	return "automatically sprints for you"
}

// Edit ...
func (a *autoSprint) Edit(p module.Property) {
	_ = a.UpdateState(p.Enabled)
}
