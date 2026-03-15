package modules

import (
	"fmt"

	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/pkg/e"
	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
)

var zoomSig = modulesutil.SignatureSettings{
	Signature: mem.MustParseSignature("F3 0F 10 89 94 00 00 00 0F 2F"),
	Patch:     mem.MustParseSignature("90 90 90 90 90 90 90 90 0F 2F"),
}

var _ module.Config = (*Zoom)(nil)

type Zoom struct {
	modulesutil.DefaultDisabled
	Process  *win.Process
	Error    func(error)
	OnToggle func(bool, e.ActionCause)
}

// Create ...
func (conf *Zoom) Create(p module.Property, cause e.ActionCause) (module.Module, error) {
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
	if cause == nil {
		cause = e.ActionCauseExternal
	}
	m := &zoom{ToggleableModule: b}
	m.Edit(p, cause)
	return m, nil
}

// Identifier ...
func (conf *Zoom) Identifier() string {
	return "zoom"
}

var _ module.Module = (*zoom)(nil)

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
func (z *zoom) Edit(p module.Property, cause e.ActionCause) {
	modulesutil.SyncState(z, p, cause)
}
