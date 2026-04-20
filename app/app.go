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
		started, init          bool
		uConfInit, modulesInit bool
		modules                *modulesMap
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
	app.data.V.uConfInit = true
	app.syncTheme(app.userConf.V)

	showErrors := app.userConf.V.ShowErrors
	app.userConf.Unlock()

	app.conf.Logger.Debug("creating module configs...")
	configs := app.setupModules(proc)
	if len(configs) == 0 {
		return errors.New("no modules created")
	}
	app.conf.Logger.Debug("created module configs.")

	app.conf.Logger.Debug("creating modules from configs...")
	modules, ok, err := app.createModulesFromConfigs(configs) //must have at least one config to call this method
	if !ok && err != nil {
		// no modules created = aborting start
		return fmt.Errorf("create modules from configs: %w", err)
	}
	if err != nil {
		app.conf.Logger.Error("error creating modules from configs", "err", err.Error())
		if showErrors {
			app.showError("create module(s) from config(s)", err)
		}
	} else {
		app.conf.Logger.Debug("created modules from configs.")
	}
	app.data.V.modules = modules
	app.data.V.modulesInit = true

	// create hotkey manager after modules initialized !!!!!
	app.userConf.Lock()
	app.hm = hotkey.ManagerConfig{Handlers: app.loadBinds(app.userConf.V)}.New()
	app.userConf.Unlock()

	app.conf.Logger.Debug("creating app content...")
	c, err := app.createContent(modules, app.win)
	if err != nil {
		return fmt.Errorf("create content: %w", err)
	}
	app.conf.Logger.Debug("created app content.")
	app.win.SetContent(c)
	return nil
}

const windowWidth, windowHeight = 450, 650

func (app *App) deployFyne() {
	app.app = fyneapp.NewWithID(ID)
	app.win = app.app.NewWindow(Name)

	app.win.SetOnClosed(func() {
		app.conf.Logger.Info("window closed via UI")
		if err := app.Close(); err != nil && !errors.Is(err, e.ErrAlreadyClosed) {
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
	done, err := app.run()
	if err != nil {
		return err
	}
	app.conf.Logger.Info("running window...")
	app.win.ShowAndRun() // blocks
	// wait for graceful window closure...
	close(done)
	app.conf.Logger.Info("window closed gracefully.")
	return nil
}

func (app *App) run() (chan struct{}, error) {
	app.data.Lock()
	defer app.data.Unlock()

	if app.data.V.started {
		return nil, e.ErrAlreadyRunning
	}
	app.conf.Logger.Info("initializing...")
	if err := app.initUnsafe(app.tr.Process()); err != nil {
		return nil, fmt.Errorf("init: %w", err)
	}
	app.conf.Logger.Info("initialized.")

	done := make(chan struct{})
	go func() {
		select {
		case <-app.ctx.Done():
			if err := app.close(closeCauseContextClosed); err != nil && !errors.Is(err, e.ErrAlreadyClosed) {
				app.conf.Logger.Error("close app", "err", err.Error())
			}
		case <-done:
		}
	}()
	app.runBackgroundTasks()
	app.data.V.started = true
	return done, nil
}

func (app *App) runBackgroundTasks() {
	go func() {
		if err := app.tr.Run(app.ctx); err != nil && !errors.Is(err, context.Canceled) {
			app.conf.Logger.Error("process tracker error", "err", err)
		}
		// it'll be closed in Close call
	}()
	app.wg.Go(func() {
		if err := app.hm.Run(app.ctx); err != nil && !errors.Is(err, context.Canceled) {
			app.conf.Logger.Error("hotkey manager error", "err", err)
		}
		// same as process tracker...
	})
}

// Close implements io.Closer.
func (app *App) Close() error {
	return app.close(nil, true)
}

var (
	closeCauseTrackerClosed = e.NewCloseCauseString("tracker closed")
	closeCauseContextClosed = e.NewCloseCause(context.Canceled)
)

// close must be called from fyne goroutine.
func (app *App) close(cause e.CloseCause, main ...bool) (multi error) {
	if err := app.closeLogic(cause); err != nil && errors.Is(err, e.ErrAlreadyClosed) {
		return err
	}
	app.conf.Logger.Debug("closing window...", "main", misc.HasTrueOption(main))
	// close window last of all !!!
	relay := func(fn func()) { fn() }
	if !misc.HasTrueOption(main) {
		relay = fyne.DoAndWait
	}
	relay(app.app.Quit) //todo: close with timeout
	return nil
}

func (app *App) closeLogic(cause e.CloseCause) error {
	if !app.closed.CompareAndSwap(false, true) {
		return e.ErrAlreadyClosed
	}
	start := time.Now()
	if cause == nil {
		cause = e.CloseCauseExternal
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

//todo: some "component" interface with methods Initialized() or PartiallyInitialized() to close all components properly even if they didnt fully
// initialized, example: first module initializes, second throws an error and the program closes. this causes to first module be unclosed

func (app *App) closeIfStarted(cause e.CloseCause) {
	app.data.Lock()
	defer app.data.Unlock() // release lock earlier because we don't want to lock until wg done

	defer func() {
		// after we disabled all modules we can close the context safely
		app.cancel()
		app.doClose("process tracker", app.tr.CloseWithProcess)
	}()
	if app.data.V.uConfInit {
		// save the user config
		app.saveUserConfig(app.app)
	}
	// if closed because parent process closed, we don't need to disable
	// all modules as it will not affect
	if app.data.V.modulesInit && !e.CloseCauseIs(cause, closeCauseTrackerClosed) {
		// disable all modules before canceling context
		// if we will not do this any logic that takes our context can end
		// earlier so it won't disable modules properly
		app.disableModulesUnsafe()
	}
}

func (app *App) disableModulesUnsafe() {
	//Disable method of module.Module must be NOT fyne-obsessed, it'll cause bugs instead
	app.conf.Logger.Debug("disabling modules...")
	defer app.conf.Logger.Debug("disabled modules.")

	for _, m := range app.data.V.modules.AllFromFront() {
		m.Disable(actionCauseModuleDisabled)
		app.conf.Logger.Debug("disabled module.", "module", m.Name())
	}
}
