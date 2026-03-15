package fyneutil

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2/widget"
	"github.com/something-that-is-cool/zutil/pkg/e"
)

type SliderWithTrackedInput struct {
	Handler        e.ErrorHandler
	Action         func(v float64, cause e.ActionCause) error
	DefaultCause   e.ActionCause
	Min, Max, Step float64
	Default        float64
	ShowRemainder  bool
	FormatFloat    func(float64) string
	Slider         *widget.Slider
	Input          *widget.Entry

	sliderRecursive bool
	inputRecursive  bool

	previousSlider float64
	previousInput  string
}

func (s *SliderWithTrackedInput) Create() (*widget.Slider, *widget.Entry) {
	if s.Action == nil {
		panic("must set action")
	}
	if s.Handler == nil {
		s.Handler = e.NopErrorHandler{}
	}
	if s.DefaultCause == nil {
		s.DefaultCause = e.ActionCauseExternal
	}
	if s.Slider == nil {
		s.Slider = widget.NewSlider(s.Min, s.Max)
		s.Slider.Value = s.Default
	}
	if s.Step <= 0 {
		s.Step = 1.0
	}
	s.Slider.Step = s.Step
	s.Slider.Min = s.Min
	s.Slider.Max = s.Max

	if s.Input == nil {
		s.Input = widget.NewEntry()
	}
	format := s.getFormatFunc()
	if s.Input.Text == "" || s.Input.Text == "0" {
		s.Input.SetText(format(s.Slider.Value))
	}
	s.Slider.OnChanged = func(f float64) {
		if s.sliderRecursive {
			return
		}
		s.inputRecursive = true
		s.Input.SetText(format(f))
		s.inputRecursive = false
	}
	s.Slider.OnChangeEnded = func(f float64) {
		if s.sliderRecursive {
			return
		}
		s.Set(f, s.DefaultCause, true)
		s.previousSlider = f
	}
	s.Input.OnChanged = func(str string) {
		if s.inputRecursive {
			return
		}
		f, err := strconv.ParseFloat(str, 64)
		if err != nil || (err == nil && (f < s.Min || f > s.Max)) {
			f = s.Slider.Value
		} else {
			s.Set(f, s.DefaultCause, true)
		}
		s.previousInput = str
		// must make input instead of slider recursive here so slider can call handler
		s.inputRecursive = true
		s.Slider.SetValue(f)
		s.Input.SetText(format(f))
		s.inputRecursive = false
	}
	s.Set(s.Slider.Value, s.DefaultCause, true)
	return s.Slider, s.Input
}

type (
	SliderInputOptionNotRefresh     struct{}
	SliderInputOptionOnlyCallAction struct{}
)

func (s *SliderWithTrackedInput) Set(v float64, cause e.ActionCause, opts ...any) {
	notRefresh, onlyCallAction := handleSliderInputOpts(opts)
	if err := s.Action(v, cause); err != nil {
		s.Handler.HandleError("slider/input action", err)
		return
	}
	if onlyCallAction {
		return
	}
	s.Slider.Value = v
	s.inputRecursive = true
	s.Input.SetText(s.getFormatFunc()(v))
	s.inputRecursive = false

	if !notRefresh {
		s.Slider.Refresh()
		s.Input.Refresh()
	}
}

func (s *SliderWithTrackedInput) getFormatFunc() func(float64) string {
	format := formatFloatDefault
	if s.ShowRemainder {
		format = formatFloatWithRemainder
	}
	if s.FormatFloat != nil {
		format = s.FormatFloat
	}
	return format
}

func formatFloatWithRemainder(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func formatFloatDefault(f float64) string {
	return fmt.Sprintf("%.0f", f)
}

func handleSliderInputOpts(x []any) (notRefresh, onlyCallAction bool) {
	for _, v := range x {
		switch v.(type) {
		case SliderInputOptionNotRefresh:
			notRefresh = true
		case SliderInputOptionOnlyCallAction:
			onlyCallAction = true
		}
	}
	return
}
