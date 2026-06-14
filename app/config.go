package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/something-that-is-cool/zutil/pkg/win"
)

type Config struct {
	Logger  *slog.Logger
	Process string
}

const GameVersion = "1.1.5"

var supportedVersion = win.MustParseVersion(GameVersion)

// New tries to create new App instance from Config, allowing to provide custom
// context to control app lifecycle.
func (conf Config) New(parent context.Context) (*App, error) {
	if conf.Process == "" {
		return nil, errors.New("empty process")
	}
	if conf.Logger == nil {
		conf.Logger = slog.Default()
	}
	proc, err := win.OpenProcess(conf.Process)
	if err != nil {
		return nil, fmt.Errorf("open process: %w", err)
	}
	if !versionSupported(proc) {
		_ = proc.Close()
		return nil, fmt.Errorf("unsupported version: %q", proc.Version())
	}
	app := &App{conf: conf}
	trackerConf := win.ProcessTrackerConfig{
		OnClose: []func(){func() {
			_ = app.close(closeCauseTrackerClosed)
		}},
		Process: proc,
	}
	app.tr, err = trackerConf.New()
	if err != nil {
		return nil, fmt.Errorf("create tracker: %w", err)
	}
	app.deployFyne()
	app.ctx, app.cancel = context.WithCancel(parent)
	return app, nil
}

func versionSupported(proc *win.Process) bool {
	return proc.Version().E(supportedVersion)
}
