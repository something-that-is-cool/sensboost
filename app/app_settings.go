package app

import (
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/internal/misc"
	"github.com/something-that-is-cool/zutil/internal/version"
	"github.com/something-that-is-cool/zutil/pkg/e"
	"github.com/something-that-is-cool/zutil/pkg/fyneutil"
)

func aboutMessage(conf Config) string {
	return misc.JoinNewLine(
		fmt.Sprintf("zutil (for %s v%s)", conf.Process, conf.SupportedVersion),
		"build "+version.Version+" "+"("+version.Commit+")",
		"using "+runtime.Version(),
		"",
		"[t.me/nigger1790]", "[t.me/zovutil]",
		"[github.com/something-that-is-cool]",
		"",
		"made by controllin",
		"turtl sosun",
		"",
		fmt.Sprintf("Copyright (C) %d Ivan Z. All rights reserved.", time.Now().Year()),
	)
}

func (app *App) createSettingsObject() fyne.CanvasObject {
	settings := &widget.Button{
		Icon:     theme.SettingsIcon(),
		OnTapped: app.showSettings,
	}
	about := &widget.Button{Icon: theme.InfoIcon(), OnTapped: func() {
		app.showInfo("About", aboutMessage(app.conf))
	}}
	toggleTheme := &widget.Button{Icon: theme.VisibilityIcon(), OnTapped: func() {
		app.userConf.Lock()
		defer app.userConf.Unlock()

		app.userConf.V.LightTheme = !app.userConf.V.LightTheme
		app.syncTheme(app.userConf.V)
	}}
	buttons := fyneutil.LeftAndRight(toggleTheme, settings)
	return fyneutil.LeftAndRight(buttons, about)
}

func (app *App) showSettings() {
	content := container.NewVBox(
		widget.NewButton("Import config", app.importConfig),
		widget.NewButton("Export config", app.exportConfig),
		widget.NewButton("Reset config", app.resetConfig),
		fyneutil.LeftAndRight(widget.NewLabel("Show errors"), app.newShowErrorsCheck()),
	)
	app.data.Lock()
	defer app.data.Unlock()

	dialog.ShowCustom("Settings", "Close", content, app.win)
}

func (app *App) importConfig() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			app.showError("import config", err)
			return
		}
		if reader == nil {
			// canceled
			return
		}
		defer reader.Close() //nolint:errcheck
		if err = app.doImport(reader); err != nil {
			app.showError("import config", err)
			return
		}
		app.showInfo("Import", misc.JoinNewLine(
			"successfully imported config.",
			reader.URI().Path(),
		))
	}, app.win)
}

var actionCauseImportedConfig = e.NewActionCause("imported config")

func (app *App) doImport(reader fyne.URIReadCloser) error {
	d, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("read all (reader): %w", err)
	}
	var conf *UserConfig
	if err = json.Unmarshal(d, &conf); err != nil {
		return fmt.Errorf("unmarshal json: %w", err)
	}
	toEdit := app.applyConfig(conf)
	app.doModuleUpdates(toEdit, actionCauseImportedConfig)
	return nil
}

func (app *App) exportConfig() {
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			app.showError("export config", err)
			return
		}
		if writer == nil {
			// canceled
			return
		}
		defer writer.Close() //nolint:errcheck
		if err = app.doExport(writer); err != nil {
			app.showError("export config", err)
			return
		}
		app.showInfo("Export config", misc.JoinNewLine(
			"successfully exported config.",
			writer.URI().Path(),
		))
	}, app.win)
}

func (app *App) doExport(writer fyne.URIWriteCloser) error {
	app.userConf.Lock()
	defer app.userConf.Unlock()
	// uConf must not be nil here

	d, err := json.MarshalIndent(app.userConf.V, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config to json: %w", err)
	}
	if _, err = writer.Write(d); err != nil {
		return fmt.Errorf("export: write config to file: %w", err)
	}
	return nil
}

var actionCauseResetConfig = e.NewActionCause("reset config")

func (app *App) resetConfig() {
	toEdit := app.applyConfig(DefaultUserConfig())
	app.doModuleUpdates(toEdit, actionCauseResetConfig)
}

var actionCauseJustCreated = e.NewActionCause("just created")

func (app *App) newShowErrorsCheck() *widget.Check {
	toggler := &fyneutil.Toggler{
		Handler: app,
		Action: func(v bool, cause e.ActionCause) error {
			app.showErrors(v, cause)
			return nil
		},
	}
	toggler.Create()
	toggler.Set(func() bool {
		app.userConf.Lock()
		defer app.userConf.Unlock()
		return app.userConf.V.ShowErrors
	}(), actionCauseJustCreated)
	return toggler.Check
}

func (app *App) showErrors(state bool, _ e.ActionCause) {
	app.userConf.Lock()
	defer app.userConf.Unlock()
	app.userConf.V.ShowErrors = state
}

func (app *App) applyConfig(newConf *UserConfig) map[module.Module]module.Property {
	app.userConf.Lock()
	defer app.userConf.Unlock()
	// uConf must not be nil here
	app.userConf.V = newConf
	// sync theme
	app.syncTheme(app.userConf.V)

	app.data.Lock()
	defer app.data.Unlock()

	func() {
		app.hm.Events.Lock()
		defer app.hm.Events.Unlock()
		// clear binds
		app.hm.ClearHandlersUnsafe()
		// apply new binds if there are
		app.applyBindsUnsafe(app.userConf.V)
	}()
	toEdit := make(map[module.Module]module.Property)
	// reset modules to default state
	for conf, m := range app.data.V.modules.AllFromFront() {
		property := conf.DefaultProperty()
		if p, ok := newConf.Modules[conf.Identifier()]; ok {
			property = p
		}
		app.userConf.V.Modules[conf.Identifier()] = property
		// schedule updating module state because it calls handler that locks userConf
		toEdit[m] = property
	}
	return toEdit
}

func (app *App) applyBindsUnsafe(conf *UserConfig) {
	for mod, char := range conf.Binds {
		m, ok := app.moduleByIDUnsafe(mod)
		if !ok {
			continue
		}
		app.hm.HandleUnsafe(char, app.bindToggleModule(mod, m))
	}
}
