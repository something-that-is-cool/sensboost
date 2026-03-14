package app

import (
	"errors"

	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/pkg/e"
	"github.com/something-that-is-cool/zutil/pkg/fyneutil"
)

func (app *App) syncTheme(conf *UserConfig) {
	variant := theme.VariantDark
	if conf.LightTheme {
		variant = theme.VariantLight
	}
	settings := app.app.Settings()
	settings.SetTheme(fyneutil.NewVariantTheme(theme.DefaultTheme(), variant))
}

func (app *App) showInfo(title, msg string) {
	dialog.ShowInformation(title, msg, app.win)
}

func (app *App) showError(err error) {
	dialog.ShowError(err, app.win)
}

func (app *App) doModuleUpdates(updates map[module.Module]module.Property) {
	for m, property := range updates {
		m.Edit(property)
	}
}

func (app *App) moduleByIDUnsafe(id string) (module.Module, bool) {
	for conf, m := range app.data.V.modules.AllFromFront() {
		if conf.Identifier() == id {
			return m, true
		}
	}
	return nil, false
}

func (app *App) doClose(name string, fn func() error) {
	app.conf.Logger.Info("closing " + name + "...")
	if err := fn(); err != nil && !errors.Is(err, e.ErrAlreadyClosed) {
		app.conf.Logger.Warn("cannot close "+name, "err", err.Error())
		return
	}
	app.conf.Logger.Info("closed " + name + ".")
}
