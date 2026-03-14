package app

import (
	_ "embed"
	"fmt"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/pkg/fyneutil"
)

func (app *App) createContent(modules *modulesMap, w fyne.Window) (fyne.CanvasObject, error) {
	f, err := app.createFooter()
	if err != nil {
		return nil, fmt.Errorf("create footer: %w", err)
	}
	var obj []fyne.CanvasObject
	for c, mod := range modules.AllFromFront() {
		window := app.createModuleWindow(mod, c, w)
		obj = append(obj, window)
	}
	a := container.NewGridWithRows(4, obj...)
	// creating scroll prevents from minimizing bug
	return container.NewVScroll(fyneutil.WithFooter(a, f)), nil
}

func (app *App) createModuleWindow(m module.Module, c module.Config, win fyne.Window) fyne.CanvasObject {
	box := container.NewVBox(m.CreateObjects()...)
	button := app.createModuleSettingsObject(m, c, win)

	stack := container.NewStack(
		box,
		fyneutil.RightBottomCorner(button),
	)
	w := container.NewInnerWindow(m.Name(), stack)
	w.CloseIntercept = func() {}
	return w
}

func (app *App) createFooter() (fyne.CanvasObject, error) {
	u, err := url.Parse("https://t.me/+rweTeGr1vOxjM2Qy")
	if err != nil {
		return nil, fmt.Errorf("parse controllin URL: %w", err)
	}
	return container.NewBorder(
		nil, nil,
		widget.NewHyperlink("Join to controllin", u),
		app.createSettingsObject(),
	), nil
}
