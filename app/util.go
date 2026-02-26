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
	app.data.app.Settings().SetTheme(fyneutil.NewVariantTheme(theme.DefaultTheme(), variant))
}

func (app *App) showInfo(title, msg string, safe bool) {
	if safe {
		app.data.Lock()
		defer app.data.Unlock()
	}
	if app.data.win == nil {
		return
	}
	dialog.ShowInformation(title, msg, app.data.win)
}

func (app *App) doModuleUpdates(updates map[module.Module]module.Property) {
	for m, property := range updates {
		m.Edit(property)
	}
}

func propertyFromValue(v any) (p module.Property) {
	if p, ok := v.(module.Property); ok {
		return p
	}
	if val, ok := v.(bool); ok {
		p.Enabled = val
		return p
	}
	p.Value = v
	return p
}
