package modules

import (
	"fmt"

	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

var zoomSig = modulesutil.SignatureSettings{
	Signature: []byte{0xF3, 0x0F, 0x10, 0x89, 0x94, 0x00, 0x00, 0x00, 0x0F, 0x2F},
	Patch:     []byte{0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x0F, 0x2F},
}

var _ module.Module = (*zoom)(nil)

type Zoom struct {
	modulesutil.DefaultDisabled
	Process  *win.Process
	Error    func(error)
	OnToggle func(bool)
}

// Create ...
func (conf *Zoom) Create(p module.Property) (module.Module, error) {
	c := &modulesutil.ByteToggleModule{
		Sig:      zoomSig,
		Process:  conf.Process,
		Error:    conf.Error,
		OnToggle: conf.OnToggle,
	}
	b, err := c.New()
	if err != nil {
		return nil, fmt.Errorf("create byte toggler module: %w", err)
	}
	m := &zoom{ToggleableModule: b}
	_ = m.UpdateState(p.Enabled)
	return m, nil
}

// Identifier ...
func (conf *Zoom) Identifier() string {
	return "zoom"
}

type zoom struct {
	modulesutil.ToggleableModule
}

// Name ...
func (z *zoom) Name() string {
	return "zoom"
}

// Description ...
func (z *zoom) Description() string {
	return "..."
}

// Edit ...
func (z *zoom) Edit(property module.Property) {
	_ = z.UpdateState(property.Enabled)
}
