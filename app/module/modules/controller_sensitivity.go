package modules

import (
	_ "embed"
	"fmt"
	"math"

	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/pkg/win"
)

var controllerSensitivityPtr = modulesutil.PointerSettings{
	BaseAddress: 0x019209F0,
	Offsets:     []uintptr{0x10, 0x8, 0x8, 0x8, 0x28, 0xB0, 0x68, 0x14},
}

var _ module.Config = (*ControllerSensitivity)(nil)

type ControllerSensitivity struct {
	Process        *win.Process
	Error          func(error)
	OnValueChanged func(float64)
}

func (conf *ControllerSensitivity) Create(p module.Property) (module.Module, error) {
	fc := &modulesutil.Float32Module{
		Process:        conf.Process,
		Error:          conf.Error,
		OnValueChanged: conf.OnValueChanged,
		Min:            1,
		Max:            300,
		Default:        100,
		SliderToMemory: func(f float64) float32 {
			return float32(math.Ceil(f)) / 100
		},
		MemoryToSlider: func(f float32) float64 {
			return math.Ceil(float64(f) * 100)
		},
		Ptr: controllerSensitivityPtr,
	}
	f, err := fc.New()
	if err != nil {
		return nil, fmt.Errorf("create float ptr module: %w", err)
	}
	c := &controllerSensitivity{ModuleWithValue: f}
	c.Edit(p)
	return c, nil
}

// DefaultProperty ...
func (*ControllerSensitivity) DefaultProperty() module.Property {
	return module.Property{
		Enabled: false,
		Value:   100.0,
	}
}

// Identifier ...
func (*ControllerSensitivity) Identifier() string {
	return "controller_sensitivity"
}

var _ module.Module = (*controllerSensitivity)(nil)

type controllerSensitivity struct {
	modulesutil.ModuleWithValue[float64]
}

// Name ...
func (*controllerSensitivity) Name() string {
	return "controller sensitivity"
}

// Description ...
func (*controllerSensitivity) Description() string {
	return "allows to modify controller sensitivity to values higher than 100"
}

// Edit ...
func (c *controllerSensitivity) Edit(p module.Property) {
	if v, ok := p.Value.(float64); ok {
		_ = c.SetValue(v)
	}
}
