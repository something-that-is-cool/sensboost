package app

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"github.com/k4ties/sensboost/app/module"
	"github.com/k4ties/sensboost/internal/pkg/win"
)

type App struct {
	ctx    context.Context
	cancel context.CancelFunc

	conf Config

	app fyne.App

	win   fyne.Window
	winMu sync.Mutex // in case of concurrent Close call

	wg sync.WaitGroup

	tr *win.ProcessTracker

	closed, started atomic.Bool

	modules   []module.Module
	modulesMu sync.Mutex // in case of concurrent Close call
}

func (app *App) init(proc *win.Process) ([]module.Module, error) {
	app.winMu.Lock()
	defer app.winMu.Unlock()

	app.app = fyneapp.New()
	app.win = app.app.NewWindow("sensboost")

	app.win.SetMaster()
	app.win.CenterOnScreen()
	app.win.Resize(fyne.NewSize(300, 250))
	app.win.SetFixedSize(true)

	app.win.SetOnClosed(func() {
		_ = app.Close(true)
	})
	c, modules, err := app.createContent(proc)
	if err != nil {
		return nil, fmt.Errorf("create content: %w", err)
	}
	app.win.SetContent(c)
	return modules, nil
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
	modules, err := app.init(app.tr.Process())
	if err != nil {
		return fmt.Errorf("init: %w", err)
	}
	app.modulesMu.Lock()
	app.modules = modules
	app.modulesMu.Unlock()
	go func() {
		<-app.ctx.Done()
		if err := app.Close(false); err != nil && !errors.Is(err, ErrAppClosed) {
			app.conf.Logger.Error("close app", "err", err.Error())
		}
	}()
	go func() {
		defer app.tr.Close()
		if err := app.tr.Run(app.ctx); err != nil {
			_ = app.Close(false)
		}
	}()
	app.win.ShowAndRun()
	return nil
}

// Close ...
func (app *App) Close(main bool) error {
	if !app.closed.CompareAndSwap(false, true) {
		return ErrAppClosed
	}
	// in case of panic do this pattern to defer unlock
	func() {
		app.modulesMu.Lock()
		defer app.modulesMu.Unlock()
		// disable all modules before canceling context
		// if we will not do this any logic that takes our context can end
		// earlier so it won't disable modules properly
		for _, m := range app.modules {
			m.Disable()
		}
	}()
	// after we disabled all modules we can close the context safely
	app.cancel()
	app.tr.Close()

	app.wg.Wait()
	if !main {
		fyne.DoAndWait(app.closeWin)
	} else {
		app.closeWin()
	}
	return nil
}

func (app *App) closeWin() {
	app.winMu.Lock()
	defer app.winMu.Unlock()

	if app.win != nil {
		app.win.Close()
	}
}
