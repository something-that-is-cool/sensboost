package app

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/internal/pkg/fyneutil"
)

func (app *App) createModuleSettingsObject(m module.Module, c module.Config, w fyne.Window) fyne.CanvasObject {
	if _, ok := m.(modulesutil.ToggleableModule); !ok {
		return fyneutil.NewClickableIcon(theme.InfoIcon(), func() {
			app.showInfo("Description", m.Description(), true)
		})
	}
	return fyneutil.NewClickableIcon(theme.SettingsIcon(), func() {
		dialog.ShowCustom("Settings", "Close", container.NewVBox(
			widget.NewButton("Description", func() {
				app.showInfo("Description", m.Description(), true)
			}),
			widget.NewButton("Bind", app.bindButton(w, m, c)),
		), w)
	})
}
