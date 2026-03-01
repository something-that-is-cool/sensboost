package app

import (
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/internal/pkg/fyneutil"
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
