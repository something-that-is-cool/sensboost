package app

import (
	"fyne.io/fyne/v2/dialog"
	"github.com/something-that-is-cool/zutil/app/module"
)

func (app *App) onError(mod string) func(err error) {
	return func(err error) {
		app.conf.Logger.Error("an error occurred", "module", mod, "err", err.Error())
	}
}

func (app *App) showInfo(title, msg string) {
	app.data.Lock()
	defer app.data.Unlock()

	if app.data.win == nil {
		return
	}
	dialog.ShowInformation(title, msg, app.data.win)
}

func (app *App) showInfoFunc(title, msg string) func() {
	return func() {
		app.showInfo(title, msg)
	}
}

func propertyFromValue(v any) (p module.Property) {
	if val, ok := v.(bool); ok {
		p.Enabled = val
	} else {
		p.Value = v
	}
	return
}
