package app

import (
	"errors"
	"fmt"

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

func (app *App) showError(title string, err error) {
	app.showInfo(fmt.Sprintf("Error: %q", title), err.Error())
}

func (app *App) doModuleUpdates(updates map[module.Module]module.Property, cause e.ActionCause) {
	for m, property := range updates {
		m.Edit(property, cause)
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
	if err := fn(); err != nil && (!errors.Is(err, e.ErrAlreadyClosed) && !errors.Is(err, e.ErrClosed)) {
		app.conf.Logger.Warn("cannot close "+name, "err", err.Error())
		return
	}
	app.conf.Logger.Info("closed " + name + ".")
}

func (app *App) ifUserConf(x func(*UserConfig) bool, fn func()) {
	app.userConf.Lock()
	res := x(app.userConf.V)
	app.userConf.Unlock()
	if res {
		fn()
	}
}

var _ e.ErrorHandler = (*App)(nil)

func (app *App) HandleError(src string, err error) {
	app.conf.Logger.Error("error handled", "src", src, "err", err.Error())
}
