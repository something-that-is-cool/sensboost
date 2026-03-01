package app

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/app/module/modules/modulesutil"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
	"github.com/something-that-is-cool/zutil/internal/pkg/win/cursor"
)

func (app *App) bindButton(w fyne.Window, m module.Module, c module.Config) func() {
	return func() {
		var custom dialog.Dialog

		scroll := container.NewVScroll(container.NewVBox(app.createBindButtons(m, c, &custom)...))
		scroll.SetMinSize(fyne.NewSize(300, 300))

		id := c.Identifier()
		custom = dialog.NewCustom("Bind "+m.Name(), "Close", container.NewVBox(
			widget.NewButton("Reset bind", app.resetBind(id, &custom)),
			widget.NewLabel("Current bind: "+app.currentBind(id)),
			scroll,
		), w)
		custom.Show()
	}
}

func (app *App) createBindButtons(m module.Module, c module.Config, d *dialog.Dialog) (b []fyne.CanvasObject) {
	for _, char := range win.AllChars {
		b = append(b, widget.NewButton(char, func() {
			app.bindModuleTo(m, c, char)()
			(*d).Dismiss()
		}))
	}
	return
}

func (app *App) bindModuleTo(m module.Module, c module.Config, char string) func() {
	return func() {
		app.userConf.Lock()
		defer app.userConf.Unlock()

		app.userConf.V.Binds[c.Identifier()] = char
		app.hm.Handle(char, app.bindToggleModule(c.Identifier(), m))
	}
}

func (app *App) currentBind(id string) string {
	app.userConf.Lock()
	defer app.userConf.Unlock()

	b, ok := app.userConf.V.Binds[id]
	if !ok {
		return "not set"
	}
	return b
}

func (app *App) resetBind(id string, d *dialog.Dialog) func() {
	return func() {
		defer (*d).Dismiss()

		app.userConf.Lock()
		defer app.userConf.Unlock()

		h, ok := app.userConf.V.Binds[id]
		if !ok {
			return
		}
		delete(app.userConf.V.Binds, id)
		app.hm.DeleteHandler(h)
	}
}

func (app *App) bindToggleModule(id string, m module.Module) func() {
	return func() {
		if !cursor.Focused() {
			return
		}
		app.data.Lock()
		defer app.data.Unlock()

		t := m.(modulesutil.ToggleableModule)
		v, _ := t.Value()

		fyne.DoAndWait(func() {
			if err := t.Set(!v); err != nil {
				app.conf.Logger.Error("cannot toggle module state", "id", id)
			}
		})
	}
}

func (app *App) loadBinds(conf *UserConfig) map[string]func() {
	x := make(map[string]func())
	for id, key := range conf.Binds {
		m, ok := app.moduleByIDUnsafe(id)
		if !ok {
			continue
		}
		if _, ok = m.(modulesutil.ToggleableModule); !ok {
			continue
		}
		x[key] = app.bindToggleModule(id, m)
	}
	return x
}
