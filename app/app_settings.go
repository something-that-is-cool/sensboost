package app

import (
	"encoding/json"
	"fmt"
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/internal/misc"
	"github.com/something-that-is-cool/zutil/internal/pkg/fyneutil"
	"github.com/something-that-is-cool/zutil/internal/pkg/version"
)

var aboutMessage = misc.JoinNewLine(
	"zutil (MC:PE 1.1.5)",
	"build "+version.Commit,
	"",
	"made by k4ties, anx1ous",
	"",
	"Copyright (C) 2026 Ivan Z. All rights reserved.",
)

func (app *App) createSettingsObject() fyne.CanvasObject {
	settings := &widget.Button{Icon: theme.SettingsIcon(), OnTapped: func() {
		app.showSettings()
	}}
	about := &widget.Button{Icon: theme.InfoIcon(), OnTapped: func() {
		app.showInfo("About", aboutMessage, true)
	}}
	return fyneutil.LeftAndRight(settings, about)
}

func (app *App) showSettings() {
	content := container.NewVBox(
		widget.NewButton("Import config", app.importConfig),
		widget.NewButton("Export config", app.exportConfig),
		widget.NewButton("Reset config", app.resetConfig),
	)
	app.data.Lock()
	defer app.data.Unlock()

	dialog.ShowCustom("Settings", "Close", content, app.data.win)
}

func (app *App) importConfig() {
	app.data.Lock()
	win := app.data.win
	app.data.Unlock()

	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(fmt.Errorf("import: %w", err), win)
			return
		}
		if reader == nil {
			// canceled
			return
		}
		defer reader.Close() //nolint:errcheck
		if err = app.doImport(reader); err != nil {
			dialog.ShowError(fmt.Errorf("import: %w", err), win)
			return
		}
		dialog.ShowInformation("Import", misc.JoinNewLine(
			"successfully imported config.",
			reader.URI().Path(),
		), win)
	}, win)
}

func (app *App) doImport(reader fyne.URIReadCloser) error {
	d, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("read all (reader): %w", err)
	}
	var conf *UserConfig
	if err = json.Unmarshal(d, &conf); err != nil {
		return fmt.Errorf("unmarshal json: %w", err)
	}
	fmt.Println("conf:", conf)
	toEdit := app.applyConfig(conf)
	app.doModuleUpdates(toEdit)
	return nil
}

func (app *App) exportConfig() {
	app.data.Lock()
	win := app.data.win
	app.data.Unlock()

	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(fmt.Errorf("export: %w", err), win)
			return
		}
		if writer == nil {
			// canceled
			return
		}
		defer writer.Close() //nolint:errcheck
		if err = app.doExport(writer); err != nil {
			dialog.ShowError(fmt.Errorf("export: %w", err), win)
			return
		}
		dialog.ShowInformation("Export config", misc.JoinNewLine(
			"successfully exported config.",
			writer.URI().Path(),
		), win)
	}, win)
}

func (app *App) doExport(writer fyne.URIWriteCloser) error {
	app.uConfMu.Lock()
	defer app.uConfMu.Unlock()
	// uConf must not be nil here

	d, err := json.MarshalIndent(app.uConf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config to json: %w", err)
	}
	if _, err = writer.Write(d); err != nil {
		return fmt.Errorf("export: write config to file: %w", err)
	}
	return nil
}

func (app *App) resetConfig() {
	toEdit := app.applyConfig(DefaultUserConfig())
	app.doModuleUpdates(toEdit)
}

func (app *App) applyConfig(newConf *UserConfig) map[module.Module]module.Property {
	app.uConfMu.Lock()
	defer app.uConfMu.Unlock()
	// uConf must not be nil here
	app.uConf = newConf
	// sync theme
	app.syncThemeUnsafe(app.uConf)

	app.data.Lock()
	defer app.data.Unlock()

	toEdit := make(map[module.Module]module.Property)
	// reset modules to default state
	for conf, m := range app.data.modules.AllFromFront() {
		property := conf.DefaultProperty()
		if p, ok := newConf.Modules[conf.Identifier()]; ok {
			property = p
		}
		app.uConf.Modules[conf.Identifier()] = property
		// schedule updating module state because it calls handler that locks uConfMu
		toEdit[m] = property
	}
	return toEdit
}
