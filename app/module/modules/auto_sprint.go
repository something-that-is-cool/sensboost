package modules

import (
	"fmt"

	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/pkg/e"
	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
)

var autoSprintSig = modulesutil.SignatureSettings{
	Signature: mem.MustParseSignature("0F B6 41 63 40 32 ED"),
	Patch:     mem.MustParseSignature("66 B8 01 00 40 30 ED"),
}

var _ module.Config = (*AutoSprint)(nil)

type AutoSprint struct {
	modulesutil.DefaultDisabled
	Process  *win.Process
	Error    func(error)
	OnToggle func(bool, e.ActionCause)
}

func (conf *AutoSprint) Create(p module.Property, cause e.ActionCause) (module.Module, error) {
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
	if cause == nil {
		cause = e.ActionCauseExternal
	}
	m := &autoSprint{ToggleableModule: b}
	m.Edit(p, cause)
	return m, nil
}

// Identifier ...
func (*AutoSprint) Identifier() string {
	return "auto_sprint"
}

var _ module.Module = (*autoSprint)(nil)

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
func (a *autoSprint) Edit(p module.Property, cause e.ActionCause) {
	modulesutil.SyncState(a, p, cause)
}
