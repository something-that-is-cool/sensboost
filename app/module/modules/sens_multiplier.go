package modules

import (
	"fmt"

	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/pkg/asm"
	"github.com/something-that-is-cool/zutil/pkg/e"
	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
)

var sensMultiplierSettings = modulesutil.Settings{
	Signature: mem.MustParseSignature("F3 0F 10 40 14 48 83 C4 10 5B C3"),
}

var _ module.Config = (*SensMultiplier)(nil)

type SensMultiplier struct {
	modulesutil.DefaultDisabled
	Process        *win.Process
	Error          func(error)
	OnToggle       func(bool, e.ActionCause)
	OnValueChanged func(float64, e.ActionCause)
}

func (conf *SensMultiplier) Create(p module.Property, cause e.ActionCause) (module.Module, error) {
	c := &modulesutil.ProxiedDetourFloatModule{
		Settings:       sensMultiplierSettings,
		Process:        conf.Process,
		TargetSize:     11,
		Error:          conf.Error,
		OnToggle:       conf.OnToggle,
		OnValueChanged: conf.OnValueChanged,
		Max:            3.0,
		Default:        1.0,
		Step:           0.1 / 2,
		ShowRemainer:   true,
		UserCode:       sensMultiplierUserCode,
	}
	b, err := c.New(cause)
	if err != nil {
		return nil, fmt.Errorf("create proxied detour float module: %w", err)
	}
	if cause == nil {
		cause = e.ActionCauseExternal
	}
	m := modulesutil.NewBaseToggleableValue(b,
		"sens multiplier",
		"multiplies camera sensitivity with the given value (affects both kbm & controller platforms)",
	)
	m.Edit(p, cause)
	return m, nil
}

func (*SensMultiplier) Identifier() string {
	return "sens_multiplier"
}

func sensMultiplierUserCode(valAddr uintptr) []byte {
	return asm.Build().
		RawStr("F3 0F 10 40 14").
		Mov64(asm.Rax, valAddr).
		MulssXmm(0, asm.Rax).
		RawStr("48 83 C4 10").
		Pop(asm.Rbx).
		Ret().
		Result()
}
