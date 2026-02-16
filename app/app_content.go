package app

import (
	"fmt"
	"math"
	"net/url"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var controllinURL = func() *url.URL {
	u, err := url.Parse("https://t.me/+rweTeGr1vOxjM2Qy")
	if err != nil {
		panic(fmt.Errorf("parse controllin link: %w", err))
	}
	return u
}()

func (app *App) createContent() (fyne.CanvasObject, error) {
	err := app.tr.ForceRead() // reading first time to init value
	if err != nil {
		return nil, fmt.Errorf("initial tracker read: %w", err)
	}
	val := normalizeSensitivity(app.tr.LastValue())

	content := createContentWidget(app, 1, 300, val)
	contentBox := container.NewVBox(content.Slider, content.Entry)

	footer := container.NewBorder(nil,
		nil,
		widget.NewHyperlink("join to controllin", controllinURL),
		widget.NewLabel("(C) Ivan Z"),
	)
	b := container.NewBorder(nil, footer, nil, nil, contentBox)
	return container.NewPadded(b), nil
}

type content struct {
	app *App

	Slider *widget.Slider
	Entry  *widget.Entry
	Value  float64

	updating bool
}

func createContentWidget(app *App, min, max, value float64) *content {
	s := &content{
		app:    app,
		Slider: widget.NewSlider(min, max),
		Entry:  widget.NewEntry(),
		Value:  value,
	}
	s.Entry.SetPlaceHolder("1...300")
	s.Slider.SetValue(value)

	s.initHandlers()
	s.setValue(value)
	return s
}

func (c *content) initHandlers() {
	c.app.tr.Handle(func(f float64) {
		fyne.Do(func() {
			c.setValue(normalizeSensitivity(f))
		})
	})
	c.Slider.OnChanged = func(value float64) {
		if c.updating {
			return
		}
		c.Value = value
		c.updateUI()
		c.writeSensitivity(value)
	}
	c.Entry.OnChanged = func(text string) {
		if c.updating || text == "" {
			return
		}
		value, err := strconv.ParseInt(text, 10, 64)
		if err != nil || float64(value) < c.Slider.Min || float64(value) > c.Slider.Max {
			c.updateUI()
			return
		}
		c.Value = float64(value)
		c.Slider.SetValue(float64(value))
		c.writeSensitivity(float64(value))
	}
}

func (c *content) setValue(value float64) {
	if value >= c.Slider.Min && value <= c.Slider.Max {
		c.updating = true
		c.Value = value
		c.Slider.SetValue(value)
		c.Entry.SetText(fmt.Sprintf("%.0f", value))
		c.updating = false
	}
}

func (c *content) updateUI() {
	c.updating = true
	c.Entry.SetText(fmt.Sprintf("%.0f", c.Value))
	c.updating = false
}

func (c *content) writeSensitivity(val float64) {
	val = math.Ceil(val) / 100
	if err := c.app.tr.WriteValue(val); err != nil {
		c.app.conf.Logger.Error("error writing new sensitivity", "err", err.Error(), "sens", val)
	}
}

func normalizeSensitivity(val float64) float64 {
	return math.Ceil(val) * 100
}
