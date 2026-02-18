package modules

import (
	_ "embed"

	"github.com/k4ties/sensboost/app/module"
	"github.com/k4ties/sensboost/app/module/modules/modulesutil"
	"github.com/k4ties/sensboost/internal/pkg/win"
)

var (
	baseAddress = uintptr(0x019209F0)
	offsets     = []uintptr{0x10, 0x8, 0x8, 0x8, 0x28, 0xB0, 0x68, 0x14}
)

var _ module.Module = (*controllerSensitivity)(nil)

type ControllerSensitivity struct {
	Process *win.Process
	Error   func(error)
}

func (conf ControllerSensitivity) Create() module.Module {
	return &controllerSensitivity{FloatPointerModule: &modulesutil.FloatPointerModule{
		Process: conf.Process,
		Error:   conf.Error,
		Min:     1,
		Max:     300,
		Default: 100,
		SliderToMemory: func(f float64) float32 {
			return float32(f) / 100
		},
		MemoryToSlider: func(f float32) float64 {
			return float64(f) * 100
		},
		BaseAddress: baseAddress,
		Offsets:     offsets,
	}}
}

type controllerSensitivity struct {
	*modulesutil.FloatPointerModule
}

// Name ...
func (*controllerSensitivity) Name() string {
	return "controller sensitivity"
}

// Description ...
func (*controllerSensitivity) Description() string {
	return "allows to modify controller sensitivity to values higher than 100"
}
