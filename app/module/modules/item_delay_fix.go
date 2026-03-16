package modules

import (
	"fmt"

	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/pkg/e"
	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
)

var itemDelayFixSignature = modulesutil.SignatureSettings{
	Signature: mem.MustParseSignature("48 89 86 88 00 00 00 48"),
	Patch:     mem.MustParseSignature("90 90 90 90 90 90 90 48"),
}

var _ module.Config = (*ItemDelayFix)(nil)

type ItemDelayFix struct {
	modulesutil.DefaultDisabled
	Process  *win.Process
	Error    func(error)
	OnToggle func(bool, e.ActionCause)
}

// Create ...
func (conf *ItemDelayFix) Create(p module.Property, cause e.ActionCause) (module.Module, error) {
	c := &modulesutil.ByteToggleModule{
		Sig:      itemDelayFixSignature,
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
	m := &itemDelayFix{ToggleableModule: b}
	m.Edit(p, cause)
	return m, nil
}

// Identifier ...
func (conf *ItemDelayFix) Identifier() string {
	return "item_delay_fix"
}

var _ module.Module = (*itemDelayFix)(nil)

type itemDelayFix struct {
	modulesutil.ToggleableModule
}

// Name ...
func (*itemDelayFix) Name() string {
	return "item delay fix"
}

// Description ...
func (*itemDelayFix) Description() string {
	return "removes attack use delay of 200 ms"
}

// Edit ...
func (i *itemDelayFix) Edit(p module.Property, cause e.ActionCause) {
	modulesutil.SyncState(i, p, cause)
}
