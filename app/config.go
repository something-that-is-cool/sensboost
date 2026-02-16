package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/k4ties/sensboost/app/sens"
)

type Config struct {
	Logger     *slog.Logger `yaml:"-"`
	Process    string       `yaml:"process"`
	BaseOffset uintptr      `yaml:"base_offset"`
	Offsets    []uintptr    `yaml:"offsets"`
}

// New tries to create new App instance from Config, allowing to provide custom
// context to control app lifecycle.
func (conf Config) New(parent context.Context) (*App, error) {
	if conf.Logger == nil {
		conf.Logger = slog.Default()
	}
	trackerConf := sens.TrackerConfig{
		ProcessName: conf.Process,
		BaseOffset:  conf.BaseOffset,
		Offsets:     conf.Offsets,
		Logger:      conf.Logger.With("from", "sens-tracker"),
	}
	tr, err := trackerConf.New()
	if err != nil {
		return nil, fmt.Errorf("create sensitivity tracker: %w", err)
	}
	ctx, cancel := context.WithCancel(parent)
	app := &App{
		ctx:    ctx,
		cancel: cancel,
		tr:     tr,
		conf:   conf,
	}
	app.wg.Add(1) //add one wg for closing
	return app, nil
}
