package modulesutil

import (
	"fmt"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
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

	OnValueChanged func(int32)
	OnStateChanged func(bool)
}

// New ...
func (conf Int32PointerNopSigToggle) New() (_ ToggleableModuleWithValue[int32], err error) {
	i := &int32PointerNopSigToggle{
		p:    conf.Ptr,
		s:    conf.Sig,
		err:  conf.Error,
		proc: conf.Process,
		//sToM: conf.SliderToMemory,
		//mToS: conf.MemoryToSlider,
		addr:    conf.FinalAddr,
		resolve: conf.ResolveAddress,
	}
	togglerConf := memutil.SignatureNopTogglerConfig{
		Process:   conf.Process,
		Signature: conf.Sig.Signature,
		NopSig:    conf.Sig.Patch,
	}
	i.toggler, err = togglerConf.New()
	if err != nil {
		return nil, fmt.Errorf("create nop sig toggler: %w", err)
	}
	i.check = &widget.Check{Text: ToggleDisabled}
	i.check.OnChanged = CheckSet(conf.Error, i.check, func(b bool, _ *widget.Check) error {
		if err := i.toggler.Set(b); err != nil {
			return fmt.Errorf("set toggler: %w", err)
		}
		// force writing value
		i.writeValue(int32(i.slider.Value))
		// and calling the handler
		conf.OnStateChanged(b)
		return nil
	})
	i.slider, i.input = fyneutil.SliderWithTrackedInput{
		Min:     float64(conf.Min),
		Max:     float64(conf.Max),
		Default: float64(conf.Default),
		OnEditSlider: func(_ *widget.Slider, _, new float64) {
			v := int32(math.Ceil(new))
			i.writeValue(v)
			conf.OnValueChanged(v)
		},
	}.Create()
	return i, nil
}

var _ ToggleableModuleWithValue[int32] = (*int32PointerNopSigToggle)(nil)

type int32PointerNopSigToggle struct {
	p PointerSettings
	s SignatureSettings

	err  func(error)
	proc *win.Process

	//sToM func(int32) int32
	//mToS func(int32) int32

	slider *widget.Slider
	input  *widget.Entry
	check  *widget.Check

	toggler *memutil.SignatureNopToggler

	addr    uintptr
	resolve func() (uintptr, error)
}

// CreateObjects ...
func (i *int32PointerNopSigToggle) CreateObjects() []fyne.CanvasObject {
	si := container.NewAdaptiveGrid(2, i.slider, i.input)
	return []fyne.CanvasObject{si, i.check}
}

// SetValue ...
func (i *int32PointerNopSigToggle) SetValue(v int32) error {
	i.slider.SetValue(float64(v))
	return nil
}

// Value ...
func (i *int32PointerNopSigToggle) Value() (int32, bool) {
	v := math.Ceil(i.slider.Value)
	return int32(v), true
}

// UpdateState ...
func (i *int32PointerNopSigToggle) UpdateState(b bool) error {
	i.check.SetChecked(b)
	return nil
}

// State ...
func (i *int32PointerNopSigToggle) State() bool {
	return i.check.Checked
}

// Disable ...
func (i *int32PointerNopSigToggle) Disable() {
	_ = i.UpdateState(false)
}

func (i *int32PointerNopSigToggle) writeValue(v int32) {
	addr, err := i.lazyAddress()
	if err != nil {
		i.err(fmt.Errorf("lazy get (resolve) ptr address: %w", err))
		return
	}
	err = mem.WriteMemory[int32](i.proc, addr, v)
	if err != nil {
		i.err(fmt.Errorf("write to pointer %d: %w", v, err))
	}
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
