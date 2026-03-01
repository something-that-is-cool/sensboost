package app

import (
	"errors"

	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/internal/pkg/fyneutil"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

func (app *App) syncThemeUnsafe(conf *UserConfig) {
	variant := theme.VariantDark
	if conf.LightTheme {
		variant = theme.VariantLight
	}
	settings := app.data.V.app.Settings()
	settings.SetTheme(fyneutil.NewVariantTheme(theme.DefaultTheme(), variant))
}

func (app *App) showInfo(title, msg string, safe bool) {
	if safe {
		app.data.Lock()
		defer app.data.Unlock()
	}
	w := app.data.V.win
	if w == nil {
		return
	}
	dialog.ShowInformation(title, msg, w)
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
	if err := fn(); err != nil && !errors.Is(err, win.ErrAlreadyClosed) {
		app.conf.Logger.Warn("cannot close "+name, "err", err.Error())
		return
	}
	app.conf.Logger.Info("closed " + name + ".")
}
