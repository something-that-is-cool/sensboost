package modules

import (
	"fmt"

	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/pkg/e"
	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
)

var noCamResetSig = modulesutil.SignatureSettings{
	Signature: mem.MustParseSignature("FF 90 ? ? ? ? ? ? ? 48 8B D6 44 8B 4C 24"), //sub_140A49260 ; call qword ptr [rax+88h]
	Patch:     mem.MustParseSignature("90 90 90 90 90 90 90"),
}

//todo improve other sigs (with ida sigmaker)

var _ module.Config = (*NoCamReset)(nil)

type NoCamReset struct {
	modulesutil.DefaultDisabled
	Process  *win.Process
	Error    func(error)
	OnToggle func(bool, e.ActionCause)
}

// Create ...
func (conf *NoCamReset) Create(p module.Property, cause e.ActionCause) (module.Module, error) {
	c := &modulesutil.ByteToggleModule{
		Sig:      noCamResetSig,
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
	n := &noCamReset{ToggleableModule: b}
	n.Edit(p, cause)
	return n, nil
}

// Identifier ...
func (conf *NoCamReset) Identifier() string {
	return "no_cam_reset"
}

var _ module.Module = (*noCamReset)(nil)

type noCamReset struct {
	modulesutil.ToggleableModule
}

// Name ...
func (*noCamReset) Name() string {
	return "no cam reset"
}

// Description ...
func (*noCamReset) Description() string {
	return "prevents teleport rotation interpolation"
}

// Edit ...
func (n *noCamReset) Edit(p module.Property, cause e.ActionCause) {
	modulesutil.SyncState(n, p, cause)
}
