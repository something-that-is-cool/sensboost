package modulesutil

import (
	"fmt"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/something-that-is-cool/zutil/pkg/e"
	"github.com/something-that-is-cool/zutil/pkg/fyneutil"
	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
	"github.com/something-that-is-cool/zutil/pkg/win/mem/memutil"
)

type Int32PointerNopSigToggle struct {
	Ptr PointerSettings
	Sig SignatureSettings

	FinalAddr      uintptr
	ResolveAddress func() (uintptr, error)

	Error   func(error)
	Process *win.Process

	Min, Max, Default int32

	//SliderToMemory func(int32) int32
	//MemoryToSlider func(int32) int32

	OnValueChanged func(int32, e.ActionCause)
	OnStateChanged func(bool, e.ActionCause)
}

// New ...
func (conf Int32PointerNopSigToggle) New() (_ ToggleableModuleWithValue[int32], err error) {
	i := &int32PointerNopSigToggle{
		ErrorHandler: errorHandler{err: conf.Error},
		p:            conf.Ptr,
		s:            conf.Sig,
		proc:         conf.Process,
		//sToM: conf.SliderToMemory,
		//mToS: conf.MemoryToSlider,
		addr:    conf.FinalAddr,
		resolve: conf.ResolveAddress,
	}
	togglerConf := memutil.SignatureNopTogglerConfig{
		Process:   conf.Process,
		Signature: conf.Sig.Signature,
		Patch:     conf.Sig.Patch,
	}
	i.toggler, err = togglerConf.New()
	if err != nil {
		return nil, fmt.Errorf("create nop sig toggler: %w", err)
	}
	i.check = &fyneutil.Toggler{
		Handler: i,
		Action: func(v bool, cause e.ActionCause) error {
			if err := i.toggler.Set(v); err != nil {
				return fmt.Errorf("set toggler: %w", err)
			}
			return i.writeValue(int32(i.si.Slider.Value), func() {
				conf.OnStateChanged(v, cause)
			})
		},
	}
	i.check.Create()
	i.si = &fyneutil.SliderWithTrackedInput{
		Min:     float64(conf.Min),
		Max:     float64(conf.Max),
		Default: float64(conf.Default),
		Action: func(newVal float64, cause e.ActionCause) error {
			v := int32(math.Ceil(newVal))
			return i.writeValue(v, func() {
				conf.OnValueChanged(v, cause)
			})
		},
	}
	i.si.Create()
	return i, nil
}

var _ ToggleableModuleWithValue[int32] = (*int32PointerNopSigToggle)(nil)

type int32PointerNopSigToggle struct {
	e.ErrorHandler

	p PointerSettings
	s SignatureSettings

	proc *win.Process

	//sToM func(int32) int32
	//mToS func(int32) int32

	si    *fyneutil.SliderWithTrackedInput
	check *fyneutil.Toggler

	toggler *memutil.SignatureNopToggler

	addr    uintptr
	resolve func() (uintptr, error)
}

// CreateObjects ...
func (i *int32PointerNopSigToggle) CreateObjects() []fyne.CanvasObject {
	si := container.NewAdaptiveGrid(2, i.si.Slider, i.si.Input)
	return []fyne.CanvasObject{si, i.check.Check}
}

// SetValue ...
func (i *int32PointerNopSigToggle) SetValue(v int32, cause e.ActionCause, opts ...any) error {
	if int32(i.si.Slider.Value) == v {
		return &e.ErrValuesIsAlready{Value: v}
	}
	i.si.Set(float64(v), cause, opts...)
	return nil
}

// Value ...
func (i *int32PointerNopSigToggle) Value() (int32, bool) {
	v := math.Ceil(i.si.Slider.Value)
	return int32(v), true
}

// UpdateState ...
func (i *int32PointerNopSigToggle) UpdateState(v bool, cause e.ActionCause, opts ...any) error {
	if i.check.Check.Checked == v {
		return &e.ErrValuesIsAlready{Value: v}
	}
	i.check.Set(v, cause, opts...)
	return nil
}

// State ...
func (i *int32PointerNopSigToggle) State() bool {
	return i.check.Check.Checked
}

// Disable ...
func (i *int32PointerNopSigToggle) Disable(cause e.ActionCause) {
	i.HandleError("disable int32 ptr module", disableOnlyAction(i, cause))
}

//todo: write only when module enables

func (i *int32PointerNopSigToggle) writeValue(v int32, after ...func()) error {
	addr, err := i.lazyAddress()
	if err != nil {
		return fmt.Errorf("lazy get (resolve) ptr address: %w", err)
	}
	err = mem.WriteMemory[int32](i.proc, addr, v)
	if err != nil {
		i.addr = 0 //force recalculate address
		return fmt.Errorf("write %d to time pointer: %w", v, err)
	}
	for _, fn := range after {
		fn()
	}
	return nil
}

func (i *int32PointerNopSigToggle) lazyAddress() (uintptr, error) {
	if i.addr != 0 {
		return i.addr, nil
	}
	if i.resolve != nil {
		a, err := i.resolve()
		if err != nil {
			return 0, err
		}
		i.addr = a
		return a, nil
	}
	addr, err := i.resolveAddress()
	if err != nil {
		return 0, fmt.Errorf("resolve pointer address: %w", err)
	}
	i.addr = addr
	return addr, nil
}

func (i *int32PointerNopSigToggle) resolveAddress() (uintptr, error) {
	return mem.ResolvePointerAddress(i.proc, i.p.BaseAddress, i.p.Offsets)
}
