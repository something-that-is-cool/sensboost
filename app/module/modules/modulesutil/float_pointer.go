package modulesutil

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/k4ties/sensboost/internal/pkg/fyneutil"
	"github.com/k4ties/sensboost/internal/pkg/win"
)

type FloatPointerModule struct { // so float64 would be DoublePointerModule
	Process *win.Process
	Error   func(error)

	Min, Max, Default float64
	SliderToMemory    func(float64) float32
	MemoryToSlider    func(float32) float64

	BaseAddress uintptr
	Offsets     []uintptr

	addr uintptr
}

// CreateObjects ...
func (m *FloatPointerModule) CreateObjects() []fyne.CanvasObject {
	v, err := m.initialRead()
	if err != nil {
		v = m.Default
		m.Error(fmt.Errorf("initial read: %w", err))
	} else {
		m.Default = v
	}
	conf := fyneutil.SliderWithTrackedInput{
		Min:     m.Min,
		Max:     m.Max,
		Default: v,
		OnEditSlider: func(_ *widget.Slider, _, new float64) {
			m.write(new)
		},
	}
	slider, input := conf.Create()
	return []fyne.CanvasObject{slider, input}
}

func (m *FloatPointerModule) Disable() {
	m.write(m.Default) // already normalizes !!!
}

func (m *FloatPointerModule) write(val float64) {
	addr, err := m.resolveAddress()
	if err != nil {
		m.Error(fmt.Errorf("write %g: %w", val, err))
		return
	}
	toWrite := m.SliderToMemory(val)
	if err = win.WriteMemory[float32](m.Process, addr, toWrite); err != nil {
		m.Error(fmt.Errorf("write %g: write memory: %w", val, err))
	}
}

func (m *FloatPointerModule) initialRead() (float64, error) {
	addr, err := m.resolveAddress()
	if err != nil {
		return 0, fmt.Errorf("resolve address: %w", err)
	}
	v, err := win.ReadMemory[float32](m.Process, addr)
	if err != nil {
		return 0, fmt.Errorf("read memory: %w", err)
	}
	// don't forget to normalize value
	return m.MemoryToSlider(v), nil
}

func (m *FloatPointerModule) resolveAddress() (uintptr, error) {
	if m.addr != 0 {
		return m.addr, nil
	}
	addr, err := win.ResolvePointerAddress(m.Process, m.Process.Module, m.BaseAddress, m.Offsets)
	if err != nil {
		return 0, err
	}
	m.addr = addr
	return addr, nil
}
