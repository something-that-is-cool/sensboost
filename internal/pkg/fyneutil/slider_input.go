package fyneutil

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2/widget"
)

type SliderWithTrackedInput struct {
	Min, Max, Default float64
	InitSlider        func(*widget.Slider)
	InitInput         func(*widget.Entry)

	OnEditSlider func(slier *widget.Slider, old, new float64)
	OnEditInput  func(input *widget.Entry, old, new string)
}

func (conf SliderWithTrackedInput) Create() (*widget.Slider, *widget.Entry) {
	if conf.OnEditSlider == nil {
		conf.OnEditSlider = func(_ *widget.Slider, _, _ float64) {}
	}
	if conf.OnEditInput == nil {
		conf.OnEditInput = func(_ *widget.Entry, _, _ string) {}
	}
	slider := widget.NewSlider(conf.Min, conf.Max)
	if conf.InitSlider != nil {
		conf.InitSlider(slider)
	}
	if slider.Value == 0 {
		slider.SetValue(conf.Default)
	}
	input := widget.NewEntry()
	if conf.InitInput != nil {
		conf.InitInput(input)
	}
	if input.Text == "" {
		input.Text = fmt.Sprint(conf.Default)
	}
	sliderRecursive := false
	inputRecursive := false

	previousSlider := 0.0
	slider.OnChanged = func(f float64) {
		if sliderRecursive {
			return
		}
		inputRecursive = true
		input.SetText(fmt.Sprint(f))
		inputRecursive = false
	}
	slider.OnChangeEnded = func(f float64) {
		if sliderRecursive {
			return
		}
		conf.OnEditSlider(slider, previousSlider, f)
		previousSlider = f
	}
	previousInput := ""
	input.OnChanged = func(s string) {
		if inputRecursive {
			return
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil || (err == nil && (f < conf.Min || f > conf.Max)) {
			f = slider.Value
		} else {
			conf.OnEditInput(input, previousInput, s)
		}
		previousInput = s
		// must make input instead of slider recursive here !!!
		inputRecursive = true
		slider.SetValue(f)
		input.SetText(fmt.Sprint(f))
		inputRecursive = false
	}
	return slider, input
}
