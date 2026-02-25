package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type UserConfig struct {
	sync.RWMutex
	Modules    map[string]any `json:"modules"`
	LightTheme bool           `json:"light_theme"`
}

func DefaultUserConfig() *UserConfig {
	return &UserConfig{
		Modules:    make(map[string]any),
		LightTheme: false,
	}
}

const ConfigFilename = "config.json"

func (app *App) loadUserConfig() (conf *UserConfig, err error) {
	if app.data.app == nil {
		return nil, errors.New("app is not initializied")
	}
	root := app.data.app.Storage().RootURI().Path()
	path := filepath.Join(root, ConfigFilename)

	d, err := os.ReadFile(path)
	switch {
	case os.IsNotExist(err):
		conf = DefaultUserConfig()
		if err = app.writeConfig(conf, path); err != nil {
			return conf, fmt.Errorf("write default config to %q: %w", path, err)
		}
		return conf, nil
	case err != nil:
		// that is the actual read error
		return conf, fmt.Errorf("read config file at %q: %w", path, err)
	}
	err = json.Unmarshal(d, &conf)
	if err != nil {
		return conf, fmt.Errorf("unmarshal config: %w", err)
	}
	if conf == nil { //nilaway
		return conf, errors.New("couldn't unmarshal config")
	}
	return conf, nil
}

// ref: github.com/gameparrot/netherconnect

func (app *App) saveUserConfig() {
	if !func() bool {
		app.data.Lock()
		defer app.data.Unlock()
		return app.data.app != nil
	}() {
		return
	}
	root := app.data.app.Storage().RootURI().Path()
	path := filepath.Join(root, ConfigFilename)

	app.conf.Logger.Debug("saving user config...", "path", path)

	app.uConf.RLock()
	defer app.uConf.RUnlock()

	if err := app.writeConfig(app.uConf, path); err != nil {
		app.conf.Logger.Error("cannot save config", "err", err.Error(), "path", path)
		return
	}
	app.conf.Logger.Debug("saved user config.", "path", path)
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
	return func(v T) {
		app.uConf.Lock()
		defer app.uConf.Unlock()
		app.uConf.Modules[id] = v
	}
}
