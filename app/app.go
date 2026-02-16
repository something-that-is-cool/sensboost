package app

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"github.com/k4ties/sensboost/app/sens"
)

type App struct {
	ctx    context.Context
	cancel context.CancelFunc

	conf Config

	app fyne.App
	win fyne.Window

	wg sync.WaitGroup

	tr *sens.Tracker

	closed, started atomic.Bool
}

func (app *App) init() (fyne.Window, error) {
	app.app = fyneapp.New()
	app.win = app.app.NewWindow("sensboost")

	app.win.SetMaster()
	app.win.CenterOnScreen()
	app.win.Resize(fyne.NewSize(300, 100))
	app.win.SetFixedSize(true)

	app.win.SetOnClosed(func() {
		_ = app.Close(true)
	})
	c, err := app.createContent()
	if err != nil {
		return nil, fmt.Errorf("create content: %w", err)
	}
	app.win.SetContent(c)
	return app.win, nil
}

var ErrAppClosed = errors.New("app closed")

var ErrAlreadyRunning = errors.New("app is already running")

// Run ...
func (app *App) Run() error {
	if app.closed.Load() {
		return ErrAppClosed
	}
	if !app.started.CompareAndSwap(false, true) {
		return ErrAlreadyRunning
	}
	if _, err := app.init(); err != nil {
		return fmt.Errorf("init: %w", err)
	}
	go func() {
		<-app.ctx.Done()
		if err := app.Close(false); err != nil && !errors.Is(err, ErrAppClosed) {
			app.conf.Logger.Error("close app", "err", err.Error())
		}
	}()
	go func() {
		defer app.tr.Close() //nolint:errcheck
		if err := app.tr.Run(app.ctx); err != nil {
			_ = app.Close(false)
		}
	}()
	app.win.ShowAndRun()
	app.wg.Wait()
	return nil
}

// Close ...
func (app *App) Close(main bool) error {
	if !app.closed.CompareAndSwap(false, true) {
		return ErrAppClosed
	}
	defer app.wg.Done()
	if !main {
		fyne.DoAndWait(app.win.Close)
	} else {
		app.win.Close()
	}
	app.cancel()
	err := app.tr.Close()
	if errors.Is(err, sens.ErrTrackerAlreadyClosed) {
		return nil
	}
	return err
}
