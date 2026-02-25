package app

import (
	"fmt"
	"net/url"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/internal/pkg/fyneutil"
)

func (app *App) createContent(modules []module.Module) (fyne.CanvasObject, error) {
	f, err := app.createFooter()
	if err != nil {
		return nil, fmt.Errorf("create footer: %w", err)
	}
	var obj []fyne.CanvasObject
	for _, mod := range modules {
		window := app.createModuleWindow(mod)
		obj = append(obj, window)
	}
	a := container.NewGridWithRows(4, obj...)
	// creating scroll prevents from minimizing bug
	return container.NewVScroll(fyneutil.WithFooter(a, f)), nil
}

func (app *App) createModuleWindow(m module.Module) fyne.CanvasObject {
	box := container.NewVBox(m.CreateObjects()...)
	button := fyneutil.NewClickableIcon(theme.InfoIcon(), app.showInfoFunc("Description", m.Description()))

	stack := container.NewStack(
		box,
		fyneutil.RightBottomCorner(button),
	)
	w := container.NewInnerWindow(m.Name(), stack)
	w.CloseIntercept = func() {
		m.Disable()
		w.Hide()
	}
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
		widget.NewLabel("Ivan Zov 2011"),
	), nil
}
