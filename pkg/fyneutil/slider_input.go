package fyneutil

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2/widget"
	"github.com/something-that-is-cool/zutil/internal/misc"
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
	}
	if s.Step <= 0 {
		s.Step = 1.0
	}
	s.Slider.Step = s.Step
	if s.Slider.Value == 0 {
		s.Slider.SetValue(s.Default)
	}
	if s.Input == nil {
		s.Input = widget.NewEntry()
	}
	format := formatFloatDefault
	if s.ShowRemainder {
		format = formatFloatWithRemainder
	}
	if s.FormatFloat != nil {
		format = s.FormatFloat
	}
	if s.Input.Text == "" {
		s.Input.Text = format(s.Default)
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
	return s.Slider, s.Input
}

func (s *SliderWithTrackedInput) Set(v float64, cause e.ActionCause, notRefresh ...bool) {
	if err := s.Action(v, cause); err != nil {
		s.Handler.HandleError("slider/input action", err)
		return
	}
	s.Slider.Value = v
	if !misc.HasTrueOption(notRefresh) {
		s.Slider.Refresh()
		s.Input.Refresh()
	}
}

func formatFloatWithRemainder(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func formatFloatDefault(f float64) string {
	return fmt.Sprintf("%.0f", f)
}
