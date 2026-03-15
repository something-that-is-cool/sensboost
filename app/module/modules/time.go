package modules

import (
	"fmt"
	"math"

	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/pkg/e"
	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
)

var (
	timePtr = modulesutil.PointerSettings{
		BaseAddress: 0x01921888,
		Offsets:     []uintptr{0x10, 0x58, 0x58, 0xD8, 0x0, 0x38, 0x194},
	}
	disableTimeSig = modulesutil.SignatureSettings{
		Signature: mem.MustParseSignature("41 89 89 94 01 00 00"),
	}
)

var _ module.Config = (*Time)(nil)

type Time struct {
	Process *win.Process
	Error   func(error)

	OnUpdateState func(bool, e.ActionCause)
	OnUpdateValue func(int32, e.ActionCause)
}

// Create ...
func (conf *Time) Create(p module.Property, cause e.ActionCause) (module.Module, error) {
	c := modulesutil.Int32PointerNopSigToggle{
		Ptr:            timePtr,
		Sig:            disableTimeSig,
		Error:          conf.Error,
		Process:        conf.Process,
		Max:            15000,
		OnValueChanged: conf.OnUpdateValue,
		OnStateChanged: conf.OnUpdateState,
	}
	i, err := c.New()
	if err != nil {
		return nil, fmt.Errorf("create i32 ptr nop sig toggle module: %w", err)
	}
	if cause == nil {
		cause = e.ActionCauseExternal
	}
	t := &time{ToggleableModuleWithValue: i}
	t.Edit(p, cause)
	return t, nil
}

// DefaultProperty ...
func (conf *Time) DefaultProperty() module.Property {
	return module.Property{Value: int32(15000)}
}

// Identifier ...
func (conf *Time) Identifier() string {
	return "time_changer"
}

var _ module.Module = (*time)(nil)

type time struct {
	modulesutil.ToggleableModuleWithValue[int32]
}

// Name ...
func (t *time) Name() string {
	return "time"
}

// Description ...
func (t *time) Description() string {
	return "allows to lock and change in-game time"
}

// Edit ...
func (t *time) Edit(p module.Property, cause e.ActionCause) {
	modulesutil.SyncState(t, p, cause)

	v, ok := getInteger(p.Value)
	if !ok {
		return
	}
	p.Value = v
	modulesutil.SyncValue[int32](t, p, cause)
}

func getInteger(val any) (int32, bool) {
	switch v := val.(type) {
	case int32:
		return v, true
	case int64:
		return int32(v), true
	case float32:
		return int32(math.Floor(float64(v))), true
	case float64:
		return int32(math.Floor(v)), true
	}
	return 0, false
}
