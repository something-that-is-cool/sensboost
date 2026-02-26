package app

import (
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/internal/pkg/fyneutil"
)

func (app *App) syncTheme(conf *UserConfig) {
	conf.RLock()
	defer conf.RUnlock()

	variant := theme.VariantDark
	if conf.LightTheme {
		variant = theme.VariantLight
	}
	app.data.app.Settings().SetTheme(fyneutil.NewVariantTheme(variant))
}

func (app *App) showInfo(title, msg string) {
	app.data.Lock()
	defer app.data.Unlock()

	if app.data.win == nil {
		return
	}
	dialog.ShowInformation(title, msg, app.data.win)
}

func propertyFromValue(v any) (p module.Property) {
	if val, ok := v.(bool); ok {
		p.Enabled = val
		return p
	}
	p.Value = v
	return p
}
