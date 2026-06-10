package modulesutil

import (
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/something-that-is-cool/zutil/pkg/win/mem"
	"github.com/something-that-is-cool/zutil/pkg/win/mem/memutil"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/internal/misc"
	"github.com/something-that-is-cool/zutil/pkg/e"
	"github.com/something-that-is-cool/zutil/pkg/fyneutil"
	"github.com/something-that-is-cool/zutil/pkg/win"
)

type ProxiedDetourFloatModule struct {
	Settings   Settings
	Address    uintptr
	TargetSize uint
	Process    *win.Process

	UserCode func(valAddr uintptr) []byte

	Min, Max, Default, Step float64
	ShowRemainer            bool

	SliderToMemory func(float64) float32
	MemoryToSlider func(float32) float64

	OnToggle       func(bool, e.ActionCause)
	OnValueChanged func(float64, e.ActionCause)
	Error          func(error)
}

func (conf ProxiedDetourFloatModule) New(_ e.ActionCause) (t ToggleableModuleWithValue[float64], err error) {
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
	if conf.OnToggle == nil {
		conf.OnToggle = func(bool, e.ActionCause) {}
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
	m.si = &fyneutil.SliderWithTrackedInput{
		Default:       conf.Default,
		Min:           conf.Min,
		Max:           conf.Max,
		Step:          conf.Step,
		ShowRemainder: conf.ShowRemainer,
		Action: func(newVal float64, cause e.ActionCause, first bool) error {
			if first || !m.State() || !m.forceWrite(newVal) {
				return nil
			}
			conf.OnValueChanged(newVal, cause)
			return nil
		},
	}
	m.si.Create()
	m.si.Slider.Disable()
	m.si.Input.Disable()

	action := func(v bool, cause e.ActionCause) error {
		if !v {
			if err := m.det.Disable(); err != nil {
				return fmt.Errorf("disable proxied detour: %w", err)
			}
			conf.OnToggle(v, cause)
			return nil
		}
		if err := m.det.Enable(m.uc, m.sToM(m.si.Slider.Value)); err != nil {
			return fmt.Errorf("enable proxied detour: %w", err)
		}
		m.val.Lock()
		m.val.V.v = m.si.Slider.Value
		m.val.V.notFirst = true
		m.val.Unlock()
		m.si.Slider.Enable()
		m.si.Input.Enable()
		conf.OnToggle(v, cause)
		return nil
	}
	m.toggler = &fyneutil.Toggler{
		Handler: m,
		Action:  action,
	}
	m.toggler.Create()
	return m, nil
}

type proxiedDetourFloatModule struct {
	e.ErrorHandler

	proc *win.Process
	det  *memutil.ProxiedDetour[float32]

	uc func(uintptr) []byte

	toggler *fyneutil.Toggler
	si      *fyneutil.SliderWithTrackedInput

	sToM func(float64) float32
	mToS func(float32) float64
	def  float64

	val misc.ValueWithRWMutex[struct {
		v        float64
		notFirst bool
	}]
}

func (m *proxiedDetourFloatModule) CreateObjects() []fyne.CanvasObject {
	return []fyne.CanvasObject{
		m.toggler.Check,
		container.NewGridWithColumns(2, m.si.Slider, m.si.Input),
	}
}

func (m *proxiedDetourFloatModule) UpdateState(v bool, cause e.ActionCause, opts ...any) error {
	if m.toggler.Check.Checked == v {
		return e.ErrValuesIsAlready{Value: v}
	}
	m.toggler.Set(v, cause, opts...)
	return nil
}

func (m *proxiedDetourFloatModule) State() bool {
	return m.toggler.Check.Checked
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

func (m *proxiedDetourFloatModule) Disable(cause e.ActionCause) {
	m.HandleError("disable proxied detour float module", disableOnlyAction(m, cause))
	if m.State() {
		m.forceWrite(m.def)
	}
}

func (m *proxiedDetourFloatModule) Edit(p module.Property, cause e.ActionCause) {
	SyncState(m, p, cause)
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
