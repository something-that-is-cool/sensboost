package modules

import (
	"fmt"
	"math"

	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

var (
	timePtr = modulesutil.PointerSettings{
		BaseAddress: 0x01921888,
		Offsets:     []uintptr{0x10, 0x58, 0x58, 0xD8, 0x0, 0x38, 0x194},
	}
	disableTimeSig = modulesutil.SignatureSettings{
		Signature: []byte{0x41, 0x89, 0x89, 0x94, 0x01, 0x00, 0x00},
	}
)

var _ module.Config = (*Time)(nil)

type Time struct {
	Process *win.Process
	Error   func(error)

	OnUpdateState func(bool)
	OnUpdateValue func(int32)
}

// Create ...
func (conf *Time) Create(p module.Property) (module.Module, error) {
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
	t := &time{ToggleableModuleWithValue: i}
	t.Edit(p)
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
func (t *time) Edit(p module.Property) {
	_ = t.UpdateState(p.Enabled)
	// don't forget to also update value
	if v, ok := getInteger(p.Value); ok {
		_ = t.SetValue(v)
	}
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
