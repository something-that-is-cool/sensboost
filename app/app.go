package app

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"github.com/something-that-is-cool/zutil/app/module"
	"github.com/something-that-is-cool/zutil/internal/pkg/win"
)

const Name = "zutil"

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
	modules := app.setupModules(proc)
	if len(modules) == 0 {
		return nil, errors.New("no modules created")
	}
	app.winMu.Lock()
	defer app.winMu.Unlock()
	app.deployFyne()

	c, err := app.createContent(modules)
	if err != nil {
		return nil, fmt.Errorf("create content: %w", err)
	}
	app.win.SetContent(c)
	return modules, nil
}

const windowWidth, windowHeight = 400, 520

func (app *App) deployFyne() {
	app.app = fyneapp.New()
	app.win = app.app.NewWindow(Name)

	app.win.SetMaster()
	app.win.CenterOnScreen()
	app.win.Resize(fyne.NewSize(windowWidth, windowHeight))
	app.win.SetFixedSize(true)
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
	app.wg.Go(func() {
		defer app.tr.Close()
		app.conf.Logger.Info("running process tracker...")

		if err := app.tr.Run(app.ctx); err != nil && !errors.Is(err, context.Canceled) {
			app.conf.Logger.Error("process tracker cannot run", "err", err.Error())
			return
		}
		app.conf.Logger.Info("process tracker ended gracefully.")
	})
	app.conf.Logger.Info("running window...")
	app.win.ShowAndRun()
	app.conf.Logger.Info("window closed.")
	return nil
}

// Close ...
func (app *App) Close(main bool) error {
	if !app.closed.CompareAndSwap(false, true) {
		return ErrAppClosed
	}
	app.conf.Logger.Info("closing app...")
	defer app.conf.Logger.Info("closed app.")

	// in case of panic do this pattern to defer unlock
	func() {
		app.conf.Logger.Debug("disabling modules...")
		defer app.conf.Logger.Debug("disabled modules.")

		app.modulesMu.Lock()
		defer app.modulesMu.Unlock()
		// disable all modules before canceling context
		// if we will not do this any logic that takes our context can end
		// earlier so it won't disable modules properly
		for _, m := range app.modules {
			m.Disable()
			app.conf.Logger.Debug("disabled module.", "module", m.Name())
		}
	}()
	// after we disabled all modules we can close the context safely
	app.cancel()
	app.tr.Close()
	// wait for additional things to end
	app.conf.Logger.Debug("waiting for waitgroup end...")
	app.wg.Wait()
	app.conf.Logger.Debug("waitgroup ended.")
	app.conf.Logger.Debug("closing window...")
	// close window last of all !!!
	if !main {
		fyne.DoAndWait(app.closeWin)
		return nil
	}
	app.closeWin()
	return nil
}

func (app *App) closeWin() {
	app.winMu.Lock()
	defer app.winMu.Unlock()

	if app.win != nil {
		app.win.Close()
		app.conf.Logger.Debug("closed window.")
		return
	}
	// make sure app.win is never set to nil after initialized
	app.conf.Logger.Error("tried to close window before it was initialized (nil)")
}

//fixme very rarely the app process can freeze forever (while the fyne window closes) why does this happen and do this happen after new update
