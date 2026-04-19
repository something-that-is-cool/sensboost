package modules

import (
	"fmt"

	"fyne.io/fyne/v2"
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/pkg/e"
	"github.com/something-that-is-cool/zutil/pkg/fyneutil"
	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/mem/memutil"
)

var (
	offVsyncOffset1 uintptr = 0x72EC22
	offVsyncOffset2         = offVsyncOffset1 + 0x10
)

type OffVsync struct {
	modulesutil.DefaultDisabled
	Process  *win.Process
	Error    func(error)
	OnToggle func(bool, e.ActionCause)
}

// Create ...
func (conf OffVsync) Create(p module.Property, cause e.ActionCause) (module.Module, error) {
	mod, err := conf.Process.GetModuleInfo()
	if err != nil {
		return nil, fmt.Errorf("get process module: %w", err)
	}
	if conf.OnToggle == nil {
		conf.OnToggle = func(bool, e.ActionCause) {}
	}
	a1 := mod.Address + offVsyncOffset1
	a2 := mod.Address + offVsyncOffset2

	t1 := &memutil.ByteToggler{
		Process:  conf.Process,
		Address:  a1,
		Original: []byte{0x1},
		Patch:    []byte{0x0},
	}
	t2 := &memutil.ByteToggler{
		Process:  conf.Process,
		Address:  a2,
		Original: []byte{0x1},
		Patch:    []byte{0x0},
	}
	t := &fyneutil.Toggler{
		Handler: modulesutil.ErrorHandler{Error: conf.Error},
		Action: func(v bool, cause e.ActionCause) error {
			if err := t1.Set(v); err != nil {
				return fmt.Errorf("patch 1: %w", err)
			}
			if err := t2.Set(v); err != nil {
				return fmt.Errorf("patch 2: %w", err)
			}
			conf.OnToggle(v, cause)
			return nil
		},
	}
	t.Create()

	o := &offVsync{
		ErrorHandler: modulesutil.ErrorHandler{Error: conf.Error},

		p1: t1,
		p2: t2,
		t:  t,
	}
	m := modulesutil.NewBaseToggleable(o,
		"off vsync",
		"force disables vertical synchronization",
	)
	m.Edit(p, cause)
	return m, nil
}

// Identifier ...
func (conf OffVsync) Identifier() string {
	return "off_vsync"
}

var _ modulesutil.ToggleableModule = (*offVsync)(nil)

type offVsync struct {
	e.ErrorHandler
	p1, p2 *memutil.ByteToggler
	t      *fyneutil.Toggler
}

// Edit ...
func (o *offVsync) Edit(p module.Property, cause e.ActionCause) {
	modulesutil.SyncState(o, p, cause)
}

// CreateObjects ...
func (o *offVsync) CreateObjects() []fyne.CanvasObject {
	return []fyne.CanvasObject{o.t.Check}
}

// Disable ...
func (o *offVsync) Disable(cause e.ActionCause) {
	o.t.Set(false, cause, fyneutil.TogglerOptionOnlyCallAction{})
}

// UpdateState ...
func (o *offVsync) UpdateState(v bool, cause e.ActionCause, opts ...any) error {
	o.t.Set(v, cause, opts...)
	return nil
}

// State ...
func (o *offVsync) State() bool {
	return o.t.Check.Checked
}
