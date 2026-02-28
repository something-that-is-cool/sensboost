package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"github.com/something-that-is-cool/zutil/app/module"
)

type UserConfig struct {
	Modules    map[string]module.Property `json:"modules"`
	LightTheme bool                       `json:"light_theme"`
}

func DefaultUserConfig() *UserConfig {
	return &UserConfig{
		Modules:    make(map[string]module.Property),
		LightTheme: false,
	}
}

const ConfigFilename = "config.json"

func (app *App) loadUserConfigUnsafe(a fyne.App) (conf *UserConfig, err error) {
	root := app.getRootPath(a)
	path := filepath.Join(root, ConfigFilename)

	_ = os.MkdirAll(filepath.Dir(path), 0755)

	d, err := os.ReadFile(path)
	switch {
	case os.IsNotExist(err):
		conf = DefaultUserConfig()
		if err = app.writeConfig(conf, path); err != nil {
			return nil, fmt.Errorf("write default config to %q: %w", path, err)
		}
		return conf, nil
	case err != nil:
		// that is the actual read error
		return nil, fmt.Errorf("read config file at %q: %w", path, err)
	}
	var c UserConfig
	if err = json.Unmarshal(d, &c); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &c, nil
}

// ref: github.com/gameparrot/netherconnect

func (app *App) saveUserConfig(a fyne.App) {
	root := app.getRootPath(a)
	path := filepath.Join(root, ConfigFilename)

	app.uConfMu.Lock()
	defer app.uConfMu.Unlock()

	if app.uConf == nil {
		// can happen because Close called concurrently
		return
	}
	if err := app.writeConfig(app.uConf, path); err != nil {
		app.conf.Logger.Error("cannot save config", "err", err.Error(), "path", path)
		return
	}
	app.conf.Logger.Debug("saved user config.", "path", path)
}

func (app *App) getRootPath(a fyne.App) string {
	root := a.Storage().RootURI().Path()
	return root
}

const filePerm = 0777

func (app *App) writeConfig(conf *UserConfig, path string) (err error) {
	var d []byte
	if d, err = json.Marshal(conf); err != nil {
		return fmt.Errorf("marshal default config: %w", err)
	}
	if err = os.WriteFile(path, d, filePerm); err != nil {
		return err
	}
	return nil
}

func (app *App) onModuleToggled(id string) func(bool) {
	return func(b bool) {
		app.editProperty(id, func(p *module.Property) {
			p.Enabled = b
		})
	}
}

func onModuleValueChanged[T any](app *App, id string) func(T) {
	return func(v T) {
		app.editProperty(id, func(p *module.Property) {
			p.Value = v
		})
	}
}

func (app *App) editProperty(id string, fn func(*module.Property)) {
	// this func is called when module created
	// module is created after uConf initialized, so it must not be nil
	app.uConfMu.Lock()
	defer app.uConfMu.Unlock()

	p := app.uConf.Modules[id]
	fn(&p)
	// update the property
	app.uConf.Modules[id] = p
}
