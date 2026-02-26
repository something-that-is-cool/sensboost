package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

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

func (app *App) loadUserConfigUnsafe() (conf *UserConfig, err error) {
	root, ok := app.getRootPath(false)
	if !ok {
		return nil, errors.New("fyne app is not initialized")
	}
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

func (app *App) saveUserConfig() {
	root, ok := app.getRootPath(true)
	if !ok {
		return
	}
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

func (app *App) getRootPath(safe bool) (string, bool) {
	if safe {
		app.data.Lock()
		defer app.data.Unlock()
	}
	if app.data.app == nil {
		return "", false
	}
	root := app.data.app.Storage().RootURI().Path()
	return root, true
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

func onModuleChange[T any](app *App, id string) func(T) {
	// this func is called when module created
	// module is created after uConf initialized, so it must not be nil
	return func(v T) {
		app.uConfMu.Lock()
		defer app.uConfMu.Unlock()

		app.uConf.Modules[id] = propertyFromValue(v)
		// log the status
		app.conf.Logger.Info("module value changed", "module", id, "new_value", v)
	}
}
