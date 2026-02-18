package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/k4ties/sensboost/internal/pkg/win"
)

type Config struct {
	Logger  *slog.Logger
	Process string
}

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
	ctx, cancel := context.WithCancel(parent)
	app := &App{
		ctx:    ctx,
		cancel: cancel,
		conf:   conf,
	}
	trackerConf := win.ProcessTrackerConfig{
		Handlers: []func(){func() {
			_ = app.Close(false)
		}},
		Process: proc,
	}
	app.tr, err = trackerConf.New()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("create tracker: %w", err)
	}
	return app, nil
}
