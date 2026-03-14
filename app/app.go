package app

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"github.com/something-that-is-cool/zutil/internal/misc"
	"github.com/something-that-is-cool/zutil/pkg/e"
	"github.com/something-that-is-cool/zutil/pkg/win"
	"github.com/something-that-is-cool/zutil/pkg/win/hotkey"
)

const (
	Name = "zutil"
	ID   = "monster.zov.zutil"
)

type App struct {
	ctx    context.Context
	cancel context.CancelFunc

	conf Config

	wg sync.WaitGroup

	tr *win.ProcessTracker
	hm *hotkey.Manager

	closed atomic.Bool

	win fyne.Window
	app fyne.App

	userConf misc.ValueWithMutex[*UserConfig]

	data misc.ValueWithMutex[struct {
		started, init bool
		modules       *modulesMap
	}]
}

func (app *App) initUnsafe(proc *win.Process) (err error) {
	if app.data.V.init {
		return e.ErrAlreadyInitialized
	}
	app.data.V.init = true // can only initialize once

	app.userConf.Lock()
	app.userConf.V, err = app.loadUserConfigUnsafe(app.app)
	if err != nil {
		app.userConf.Unlock()
		return fmt.Errorf("load user config: %w", err)
	}
	app.syncTheme(app.userConf.V)
	app.userConf.Unlock()

	configs := app.setupModules(proc)
	if len(configs) == 0 {
		return errors.New("no modules created")
	}
	modules, ok, err := app.createModulesFromConfigs(configs)
	if !ok && err != nil {
		// no modules created = aborting start
		return fmt.Errorf("create modules from configs: %w", err)
	}
	if err != nil {
		app.conf.Logger.Error("error creating modules from configs", "err", err.Error())
	}
	app.data.V.modules = modules

	// create hotkey manager after modules initialized !!!!!
	app.userConf.Lock()
	app.hm = hotkey.ManagerConfig{Handlers: app.loadBinds(app.userConf.V)}.New()
	app.userConf.Unlock()

	c, err := app.createContent(modules, app.win)
	if err != nil {
		return fmt.Errorf("create content: %w", err)
	}
	app.win.SetContent(c)
	return nil
}

const windowWidth, windowHeight = 450, 550

func (app *App) deployFyne() {
	app.app = fyneapp.NewWithID(ID)
	app.win = app.app.NewWindow(Name)

	app.win.SetOnClosed(func() {
		app.conf.Logger.Info("window closed via UI")
		if err := app.Close(); err != nil {
			app.conf.Logger.Error("close app via ui", "err", err.Error())
		}
	})
	app.win.SetMaster()
	app.win.CenterOnScreen()
	app.win.Resize(fyne.NewSize(windowWidth, windowHeight))
	app.win.SetFixedSize(true)
}

// Run ...
func (app *App) Run() error {
	if app.closed.Load() {
		return e.ErrClosed
	}
	app.data.Lock()
	if app.data.V.started {
		app.data.Unlock()
		return e.ErrAlreadyRunning
	}
	if err := app.initUnsafe(app.tr.Process()); err != nil {
		app.data.Unlock()
		return fmt.Errorf("init: %w", err)
	}
	go func() {
		<-app.ctx.Done()
		if err := app.closeLogic(closeCauseContextClosed); err != nil && !errors.Is(err, e.ErrClosed) {
			app.conf.Logger.Error("close app logic", "err", err.Error())
		}
	}()
	app.runBackgroundTasks()

	app.data.V.started = true
	app.data.Unlock()

	app.conf.Logger.Info("running window...")
	app.win.ShowAndRun() // blocks
	// wait for graceful window closure...
	app.conf.Logger.Info("window closed gracefully.")
	return nil
}

func (app *App) runBackgroundTasks() {
	go func() {
		defer app.tr.Close()
		if err := app.tr.Run(app.ctx); err != nil && !errors.Is(err, context.Canceled) {
			app.conf.Logger.Error("process tracker error", "err", err)
		}
	}()
	app.wg.Go(func() {
		defer app.doClose("hotkey manager", app.hm.Close)
		if err := app.hm.Run(app.ctx); err != nil && !errors.Is(err, context.Canceled) {
			app.conf.Logger.Error("hotkey manager error", "err", err)
		}
	})
}

// Close implements io.Closer.
func (app *App) Close() error {
	return app.close(nil)
}

// close must be called from fyne goroutine.
func (app *App) close(cause error) (multi error) {
	if err := app.closeLogic(cause); err != nil {
		// already closed
		return err
	}
	app.conf.Logger.Debug("closing window...")
	// close window last of all !!!
	app.app.Quit()
	return nil
}

func (app *App) closeLogic(cause error) error {
	if !app.closed.CompareAndSwap(false, true) {
		return e.ErrClosed
	}
	start := time.Now()
	if cause == nil {
		cause = closeCauseExternal
	}
	app.conf.Logger.Info("closing app logic...", "cause", cause.Error())
	defer func() {
		app.conf.Logger.Info("closed app logic.", "elapsed", time.Since(start).String())
	}()
	app.closeIfStarted(cause)
	// wait for additional things to end
	app.conf.Logger.Debug("waiting for waitgroup end...")
	app.wg.Wait()
	app.conf.Logger.Debug("waitgroup ended.")
	return nil
}

func (app *App) closeIfStarted(cause error) {
	app.data.Lock()
	defer app.data.Unlock() // release lock earlier because we don't want to lock until wg done

	if app.data.V.started {
		// save the user config
		app.saveUserConfig(app.app)
		// if closed because parent process closed, we don't need to disable
		// all modules as it will not affect
		if !errors.Is(cause, closeCauseTrackerClosed) {
			// disable all modules before canceling context
			// if we will not do this any logic that takes our context can end
			// earlier so it won't disable modules properly
			app.disableModulesUnsafe()
		}
	}
	// after we disabled all modules we can close the context safely
	app.cancel()
	app.tr.Close()
}

func (app *App) disableModulesUnsafe() {
	app.conf.Logger.Debug("disabling modules...")
	defer app.conf.Logger.Debug("disabled modules.")

	for _, m := range app.data.V.modules.AllFromFront() {
		m.Disable()
		app.conf.Logger.Debug("disabled module.", "module", m.Name())
	}
}
