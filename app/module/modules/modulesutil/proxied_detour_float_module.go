package modulesutil

import (
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/internal/misc"
	"github.com/something-that-is-cool/zutil/pkg/e"
	"github.com/something-that-is-cool/zutil/pkg/fyneutil"
	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
	"github.com/something-that-is-cool/zutil/pkg/win/mem/memutil"
)

type ProxiedDetourFloatModule struct {
	Settings   Settings
	Address    uintptr
	TargetSize uint
	Process    *win.Process

	UserCode memutil.UserCodeFunc

	Min, Max, Default, Step float64
	ShowRemainer            bool

	SliderToMemory func(float64) float32
	MemoryToSlider func(float32) float64

	OnValueChanged func(float64, e.ActionCause)
	Error          func(error)
}

func (conf ProxiedDetourFloatModule) New(initialCause e.ActionCause) (t ModuleWithValue[float64], err error) {
	if conf.UserCode == nil {
		return nil, errors.New("empty user code")
	}
	if conf.Address <= 0 && conf.Settings.Signature.Empty() {
		return nil, errors.New("empty address")
	}
	if conf.SliderToMemory == nil {
		conf.SliderToMemory = func(f float64) float32 { return float32(f) }
	}
	if conf.MemoryToSlider == nil {
		conf.MemoryToSlider = func(f float32) float64 { return float64(f) }
	}
	if conf.OnValueChanged == nil {
		conf.OnValueChanged = func(float64, e.ActionCause) {}
	}
	addr := conf.Address
	if addr <= 0 {
		addr, err = mem.ScanSignature(conf.Process, conf.Settings.Signature)
		if err != nil {
			return nil, fmt.Errorf("scan sig: %w", err)
		}
	}
	m := &proxiedDetourFloatModule{
		ErrorHandler: ErrorHandler{Error: conf.Error},
		proc:         conf.Process,
		det:          memutil.NewProxiedDetour[float32](conf.Process, addr+conf.Settings.Offset, conf.TargetSize),
		uc:           conf.UserCode,
		sToM:         conf.SliderToMemory,
		mToS:         conf.MemoryToSlider,
		def:          conf.Default,
	}
	if err := m.det.Enable(m.uc, m.sToM(conf.Default)); err != nil {
		return nil, fmt.Errorf("enable proxied detour: %w", err)
	}
	m.si = &fyneutil.SliderWithTrackedInput{
		Default:       conf.Default,
		Min:           conf.Min,
		Max:           conf.Max,
		Step:          conf.Step,
		ShowRemainder: conf.ShowRemainer,
		Action: func(newVal float64, cause e.ActionCause, first bool) error {
			if !m.forceWrite(newVal) {
				return nil
			}
			if first {
				cause = initialCause
			}
			conf.OnValueChanged(newVal, cause)
			return nil
		},
	}
	m.si.Create()
	return m, nil
}

var _ ModuleWithValue[float64] = (*proxiedDetourFloatModule)(nil)

type proxiedDetourFloatModule struct {
	e.ErrorHandler

	proc *win.Process
	det  *memutil.ProxiedDetour[float32]

	uc memutil.UserCodeFunc

	si *fyneutil.SliderWithTrackedInput

	sToM func(float64) float32
	mToS func(float32) float64
	def  float64

	val misc.ValueWithRWMutex[struct {
		v        float64
		notFirst bool
	}]
}

func (m *proxiedDetourFloatModule) CreateObjects() []fyne.CanvasObject {
	return []fyne.CanvasObject{m.si.Slider, m.si.Input}
}

func (m *proxiedDetourFloatModule) SetValue(v float64, cause e.ActionCause, opts ...any) error {
	if mgl64.FloatEqual(m.si.Slider.Value, v) {
		return e.ErrValuesIsAlready{Value: v}
	}
	if cause == nil {
		cause = e.ActionCauseExternal
	}
	m.si.Set(v, cause, opts...)
	return nil
}

func (m *proxiedDetourFloatModule) Value() (float64, bool) {
	m.val.RLock()
	defer m.val.RUnlock()

	if !m.val.V.notFirst {
		return 0, false
	}
	return m.val.V.v, true
}

func (m *proxiedDetourFloatModule) Disable(_ e.ActionCause) {
	m.forceWrite(m.def)
	if err := m.det.Disable(); err != nil {
		m.HandleError("disable underlying detour", err)
	}
}

func (m *proxiedDetourFloatModule) Edit(p module.Property, cause e.ActionCause) {
	SyncValue[float64](m, p, cause)
}

func (m *proxiedDetourFloatModule) write(val float64) error {
	m.val.Lock()
	defer m.val.Unlock()

	if m.val.V.notFirst && mgl64.FloatEqual(m.val.V.v, val) {
		return e.ErrValuesIsAlready{Value: val}
	}
	if err := m.det.WriteValue(m.sToM(val)); err != nil {
		return fmt.Errorf("write proxy memory: %w", err)
	}
	m.val.V.v = val
	m.val.V.notFirst = true
	return nil
}

func (m *proxiedDetourFloatModule) forceWrite(val float64) bool {
	if err := m.write(val); err != nil {
		m.HandleError(fmt.Sprintf("write %g to proxy", val), err)
		return false
	}
	return true
}
