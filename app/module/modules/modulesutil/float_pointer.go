package modulesutil

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/something-that-is-cool/zutil/internal/misc"
	"github.com/something-that-is-cool/zutil/internal/pkg/fyneutil"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

type FloatPointerModule struct { // so float64 would be DoublePointerModule
	Process *win.Process
	Error   func(error)

	Min, Max, Default float64
	SliderToMemory    func(float64) float32
	MemoryToSlider    func(float32) float64

	Ptr PointerSettings

	OnValueChanged func(float64)
}

// New ...
func (conf FloatPointerModule) New() (ModuleWithValue[float64], error) {
	f := &floatPointerModule{
		err:  conf.Error,
		proc: conf.Process,
		sToM: conf.SliderToMemory,
		mToS: conf.MemoryToSlider,
		min:  conf.Min,
		max:  conf.Max,
		def:  conf.Default,
		ptr:  conf.Ptr,
	}
	v, err := f.initialRead()
	if err != nil {
		v = conf.Default
		conf.Error(fmt.Errorf("initial read: %w", err))
	}
	c := fyneutil.SliderWithTrackedInput{
		Min:     conf.Min,
		Max:     conf.Max,
		Default: v,
		OnEditSlider: func(_ *widget.Slider, _, new float64) {
			if !f.forceWrite(new) {
				return
			}
			conf.OnValueChanged(new)
		},
	}
	f.slider, f.input = c.Create()
	return f, nil
}

var _ ModuleWithValue[float64] = (*floatPointerModule)(nil)

type floatPointerModule struct {
	err  func(error)
	proc *win.Process

	sToM func(float64) float32
	mToS func(float32) float64

	min, max, def float64

	ptr PointerSettings

	slider *widget.Slider
	input  *widget.Entry

	val misc.ValueWithRWMutex[struct {
		v        float64
		notFirst bool
	}]
	a uintptr
}

// CreateObjects ...
func (m *floatPointerModule) CreateObjects() []fyne.CanvasObject {
	return []fyne.CanvasObject{m.slider, m.input}
}

// SetValue ...
func (m *floatPointerModule) SetValue(v float64) error {
	m.slider.SetValue(v)
	return nil
}

// Value ...
func (m *floatPointerModule) Value() (float64, bool) {
	m.val.RLock()
	defer m.val.RUnlock()

	if !m.val.V.notFirst {
		return 0, false
	}
	return m.val.V.v, true
}

// Disable ...
func (m *floatPointerModule) Disable() {
	m.forceWrite(m.def) // already normalizes !!!
}

func (m *floatPointerModule) write(val float64) error {
	m.val.Lock()
	defer m.val.Unlock()

	if m.val.V.notFirst && mgl64.FloatEqual(m.val.V.v, val) {
		// new value is same as current
		return fmt.Errorf("new value is same as current (%g)", val)
	}
	addr, err := m.resolveAddress()
	if err != nil {
		return fmt.Errorf("resolve address: %w", err)
	}
	toWrite := m.sToM(val)
	if err = win.WriteMemory[float32](m.proc, addr, toWrite); err != nil {
		return fmt.Errorf("write memory: %w", err)
	}
	m.val.V.v = val
	m.val.V.notFirst = true
	return nil
}

func (m *floatPointerModule) forceWrite(val float64) bool {
	err := m.write(val)
	if err != nil {
		m.err(fmt.Errorf("write %g: %w", val, err))
		return false
	}
	return true
}

func (m *floatPointerModule) initialRead() (float64, error) {
	addr, err := m.resolveAddress()
	if err != nil {
		return 0, fmt.Errorf("resolve address: %w", err)
	}
	v, err := win.ReadMemory[float32](m.proc, addr)
	if err != nil {
		return 0, fmt.Errorf("read memory: %w", err)
	}
	// don't forget to normalize value
	return m.mToS(v), nil
}

func (m *floatPointerModule) resolveAddress() (uintptr, error) {
	if m.a != 0 {
		return m.a, nil
	}
	addr, err := win.ResolvePointerAddress(m.proc, m.proc.Module, m.ptr.BaseAddress, m.ptr.Offsets)
	if err != nil {
		return 0, err
	}
	m.a = addr
	return addr, nil
}
