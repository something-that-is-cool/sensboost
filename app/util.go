package app

import "fyne.io/fyne/v2/dialog"

func (app *App) onError(mod string) func(err error) {
	return func(err error) {
		app.conf.Logger.Error("an error occurred", "module", mod, "err", err.Error())
	}
}

func (app *App) showInfo(title, msg string) {
	dialog.ShowInformation(title, msg, app.win)
}

func (app *App) showInfoFunc(title, msg string) func() {
	return func() {
		app.showInfo(title, msg)
	}
}
